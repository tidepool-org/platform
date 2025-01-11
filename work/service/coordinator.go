package service

import (
	"context"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/work"
)

const (
	CoordinatorFrequencyDefault = 5 * time.Minute
	CoordinatorDelayJitter      = 0.1
)

type WorkClient interface {
	Poll(ctx context.Context, poll *work.Poll) ([]*work.Work, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *work.Update) (*work.Work, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (*work.Work, error)
}

type Coordinator struct {
	logger                   log.Logger
	workClient               WorkClient
	processors               map[string]work.Processor
	typeQuantities           work.TypeQuantities
	frequency                time.Duration
	workersCompletionChannel chan coordinatorProcessingCompletion
	workersContext           context.Context
	workersCancelFunc        context.CancelFunc
	workersWaitGroup         sync.WaitGroup
	managerContext           context.Context
	managerCancelFunc        context.CancelFunc
	managerWaitGroup         sync.WaitGroup
	timer                    *time.Timer
}

func NewCoordinator(logger log.Logger, workClient WorkClient) (*Coordinator, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if workClient == nil {
		return nil, errors.New("work client is missing")
	}

	return &Coordinator{
		logger:         logger,
		workClient:     workClient,
		processors:     map[string]work.Processor{},
		typeQuantities: work.NewTypeQuantities(),
		frequency:      CoordinatorFrequencyDefault,
	}, nil
}

func (c *Coordinator) RegisterProcessors(processors []work.Processor) error {
	for _, processor := range processors {
		if err := c.RegisterProcessor(processor); err != nil {
			return err
		}
	}
	return nil
}

func (c *Coordinator) RegisterProcessor(processor work.Processor) error {
	if processor == nil {
		return errors.New("processor is missing")
	}

	processorType := processor.Type()
	if processorType == "" {
		return errors.New("processor type is empty")
	}
	processorQuantity := processor.Quantity()
	if processorQuantity <= 0 {
		return errors.New("processor quantity is invalid")
	}
	processorFrequency := processor.Frequency()
	if processorFrequency <= 0 {
		return errors.New("processor frequency is invalid")
	}

	if c.workersCompletionChannel != nil {
		return errors.New("coordinator already started")
	}

	c.processors[processorType] = processor
	c.typeQuantities.Set(processorType, processorQuantity)
	if c.frequency > processorFrequency {
		c.frequency = processorFrequency
	}

	return nil
}

func (c *Coordinator) Start() {
	if c.workersCompletionChannel != nil {
		return
	}

	c.workersCompletionChannel = make(chan coordinatorProcessingCompletion, c.typeQuantities.Total())

	commonContext := log.NewContextWithLogger(context.Background(), c.logger)

	workersContext, workersCancelFunc := context.WithCancel(commonContext)
	c.workersContext = workersContext
	c.workersCancelFunc = workersCancelFunc

	managerContext, managerCancelFunc := context.WithCancel(commonContext)
	c.managerContext = managerContext
	c.managerCancelFunc = managerCancelFunc

	c.startManager()
}

func (c *Coordinator) Stop() {
	if c.workersCompletionChannel == nil {
		return
	}

	c.workersCancelFunc()
	c.workersWaitGroup.Wait()
	c.workersCancelFunc = nil
	c.workersContext = nil

	close(c.workersCompletionChannel)

	c.managerCancelFunc()
	c.managerWaitGroup.Wait()
	c.managerCancelFunc = nil
	c.managerContext = nil

	c.workersCompletionChannel = nil
}

func (c *Coordinator) startManager() {
	c.managerWaitGroup.Add(1)

	go func() {
		defer c.managerWaitGroup.Done()

		c.startTimer()
		defer c.stopTimer()

		for {
			select {
			case <-c.managerContext.Done(): // Drain and complete any interrupted tasks
				for completion := range c.workersCompletionChannel {
					c.completeWork(completion)
				}
				return
			case completion := <-c.workersCompletionChannel:
				c.stopTimer()
				c.completeWork(completion)
				c.requestAndDispatchWork()
				c.startTimer()
			case <-c.timer.C:
				c.requestAndDispatchWork()
				c.startTimer()
			}
		}
	}()
}

func (c *Coordinator) requestAndDispatchWork() {
	if typeQuantities := c.typeQuantities.NonZero(); !typeQuantities.IsEmpty() {
		poll := work.NewPoll()
		poll.TypeQuantities = typeQuantities
		if wrks, err := c.workClient.Poll(c.managerContext, poll); err != nil {
			log.LoggerFromContext(c.managerContext).WithError(err).Error("Failure polling for work")
		} else {
			for _, wrk := range wrks {
				c.dispatchWork(wrk)
			}
		}
	}
}

func (c *Coordinator) dispatchWork(wrk *work.Work) {
	ctx := log.ContextWithField(c.workersContext, "work", wrk)

	c.typeQuantities.Decrement(wrk.Type)
	c.workersWaitGroup.Add(1)
	go func() {
		defer c.workersWaitGroup.Done()
		defer func() {
			if err := recover(); err != nil {
				log.LoggerFromContext(ctx).WithFields(log.Fields{"error": err, "stack": string(debug.Stack())}).Error("Unhandled panic")
			}
		}()
		c.workersCompletionChannel <- c.processWork(ctx, wrk)
	}()
}

func (c *Coordinator) processWork(ctx context.Context, wrk *work.Work) coordinatorProcessingCompletion {
	identifier := &coordinatorProcessingIdentifier{
		ID:       wrk.ID,
		Type:     wrk.Type,
		Revision: wrk.Revision,
	}
	updater := &coordinatorProcessingUpdater{
		WorkClient: c.workClient,
		Identifier: identifier,
	}
	completion := coordinatorProcessingCompletion{
		Identifier: identifier,
	}

	processor, ok := c.processors[wrk.Type]
	if !ok {
		log.LoggerFromContext(ctx).Error("Processor not found for work type")
		return completion
	}

	// If the work has a processing timeout time specified
	if wrk.ProcessingTimeoutTime != nil {

		// Cancel context at processing timeout time
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, *wrk.ProcessingTimeoutTime)
		defer cancel()

		// Log if past processing timeout time
		defer func() {
			if time.Now().After(*wrk.ProcessingTimeoutTime) {
				log.LoggerFromContext(ctx).Warn("Processing duration exceeds timeout time")
			}
		}()
	}

	// Process the work, allowing for intermediate updates, and return completion
	completion.PendingUpdate = processor.Process(ctx, wrk, updater)
	return completion
}

func (c *Coordinator) completeWork(completion coordinatorProcessingCompletion) {
	ctx, lgr := log.ContextAndLoggerWithField(c.workersContext, "completion", completion)

	c.typeQuantities.Increment(completion.Identifier.Type)

	if completion.PendingUpdate != nil {
		condition := &request.Condition{
			Revision: &completion.Identifier.Revision,
		}
		update := &work.Update{
			PendingUpdate: completion.PendingUpdate,
			StateUpdate: &work.StateUpdate{
				State: work.StatePending,
			},
		}
		if wrk, err := c.workClient.Update(ctx, completion.Identifier.ID, condition, update); err != nil {
			lgr.WithError(err).Error("Failure to update work when processing complete")
		} else {
			lgr.WithField("work", wrk).Info("Updated work when processing complete")
		}
	} else {
		if _, err := c.workClient.Delete(ctx, completion.Identifier.ID, nil); err != nil {
			lgr.WithError(err).Error("Failure to delete work when processing complete")
		}
	}
}

func (c *Coordinator) startTimer() {
	jitter := int64(float64(c.frequency) * CoordinatorDelayJitter)
	frequencyWithJitter := c.frequency + time.Duration(rand.Int63n(jitter*2+1)-jitter)
	if c.timer == nil {
		c.timer = time.NewTimer(frequencyWithJitter)
	} else {
		c.timer.Reset(frequencyWithJitter)
	}
}

func (c *Coordinator) stopTimer() {
	if c.timer != nil {
		if !c.timer.Stop() {
			<-c.timer.C
		}
		c.timer = nil
	}
}

type coordinatorProcessingIdentifier struct {
	ID       string `json:"id,omitempty"`
	Type     string `json:"type,omitempty"`
	Revision int    `json:"revision,omitempty"`
}

type coordinatorProcessingUpdater struct {
	WorkClient WorkClient                       `json:"-"`
	Identifier *coordinatorProcessingIdentifier `json:"identifier,omitempty"`
}

func (c *coordinatorProcessingUpdater) ProcessingUpdate(ctx context.Context, processingUpdate *work.ProcessingUpdate) (*work.Work, error) {
	condition := &request.Condition{Revision: &c.Identifier.Revision}
	update := &work.Update{ProcessingUpdate: processingUpdate}
	wrk, err := c.WorkClient.Update(context.WithoutCancel(ctx), c.Identifier.ID, condition, update)
	if err == nil {
		c.Identifier.Revision = wrk.Revision
	}
	return wrk, err
}

type coordinatorProcessingCompletion struct {
	Identifier    *coordinatorProcessingIdentifier `json:"identifier,omitempty"`
	PendingUpdate *work.PendingUpdate              `json:"pendingUpdate,omitempty"`
}
