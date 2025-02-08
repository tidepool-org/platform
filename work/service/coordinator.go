package service

import (
	"context"
	"math/rand"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/work"
)

const (
	CoordinatorFrequencyDefault = 5 * time.Minute
	CoordinatorDelayJitter      = 0.1
)

type ServerSessionTokenProvider interface {
	ServerSessionToken() (string, error)
}

type WorkClient interface {
	Poll(ctx context.Context, poll *work.Poll) ([]*work.Work, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *work.Update) (*work.Work, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (*work.Work, error)
}

type Coordinator struct {
	logger                     log.Logger
	serverSessionTokenProvider ServerSessionTokenProvider
	workClient                 WorkClient
	processors                 map[string]work.Processor
	typeQuantities             work.TypeQuantities
	frequency                  time.Duration
	workersCompletionChannel   chan *coordinatorProcessingCompletion
	workersContext             context.Context
	workersCancelFunc          context.CancelFunc
	workersWaitGroup           sync.WaitGroup
	managerContext             context.Context
	managerCancelFunc          context.CancelFunc
	managerWaitGroup           sync.WaitGroup
	timer                      *time.Timer
}

func NewCoordinator(logger log.Logger, serverSessionTokenProvider ServerSessionTokenProvider, workClient WorkClient) (*Coordinator, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if serverSessionTokenProvider == nil {
		return nil, errors.New("server session token provider is missing")
	}
	if workClient == nil {
		return nil, errors.New("work client is missing")
	}

	return &Coordinator{
		logger:                     logger,
		serverSessionTokenProvider: serverSessionTokenProvider,
		workClient:                 workClient,
		processors:                 map[string]work.Processor{},
		typeQuantities:             work.NewTypeQuantities(),
		frequency:                  CoordinatorFrequencyDefault,
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

	c.workersCompletionChannel = make(chan *coordinatorProcessingCompletion, c.typeQuantities.Total())

	commonContext := log.NewContextWithLogger(context.Background(), c.logger)

	workersContext, workersCancelFunc := context.WithCancel(commonContext)
	workersContext = auth.NewContextWithServerSessionTokenProvider(workersContext, c.serverSessionTokenProvider)
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
	ctx := c.workersContext
	if ctx == nil {
		return
	}

	typeQuantities := c.typeQuantities.NonZero()
	if typeQuantities.IsEmpty() {
		return
	}

	poll := &work.Poll{TypeQuantities: typeQuantities}
	wrks, err := c.workClient.Poll(context.WithoutCancel(c.managerContext), poll)
	if err != nil {
		log.LoggerFromContext(c.managerContext).WithError(err).Error("Failure polling for work")
		return
	}

	for _, wrk := range wrks {
		c.dispatchWork(log.ContextWithField(ctx, "work", wrk), wrk)
	}
}

func (c *Coordinator) dispatchWork(ctx context.Context, wrk *work.Work) {
	c.typeQuantities.Decrement(wrk.Type)
	c.workersWaitGroup.Add(1)
	go func() {
		defer c.workersWaitGroup.Done()
		c.workersCompletionChannel <- c.processWork(ctx, wrk)
	}()
}

func (c *Coordinator) processWork(ctx context.Context, wrk *work.Work) *coordinatorProcessingCompletion {
	completion := &coordinatorProcessingCompletion{
		Identifier: &coordinatorProcessingIdentifier{
			ID:       wrk.ID,
			Type:     wrk.Type,
			Revision: wrk.Revision,
		},
	}
	c.processWorkWithCompletion(ctx, wrk, completion)
	return completion
}

func (c *Coordinator) processWorkWithCompletion(ctx context.Context, wrk *work.Work, completion *coordinatorProcessingCompletion) {
	defer func() {
		if err := recover(); err != nil {
			stack := strings.Split(strings.ReplaceAll(string(debug.Stack()), "\t", ""), "\n")
			log.LoggerFromContext(ctx).WithFields(log.Fields{"error": err, "stack": stack}).Error("Unhandled panic")
			completion.ProcessResult = work.NewProcessResultFailing(work.FailingUpdate{
				FailingError:      errors.Serializable{Error: errors.WithMeta(errors.Newf("unhandled panic: %v", err), stack)},
				FailingRetryCount: 1,
				FailingRetryTime:  time.Now().Add(5 * time.Second),
				Metadata:          wrk.Metadata,
			})
		}
	}()

	processor, ok := c.processors[wrk.Type]
	if !ok {
		completion.ProcessResult = work.NewProcessResultFailed(work.FailedUpdate{
			FailedError: errors.Serializable{Error: errors.New("processor not found for type")},
			Metadata:    wrk.Metadata,
		})
		return
	}

	updater := &coordinatorProcessingUpdater{
		WorkClient: c.workClient,
		Identifier: completion.Identifier,
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

	completion.ProcessResult = processor.Process(ctx, wrk, updater)
}

func (c *Coordinator) completeWork(completion *coordinatorProcessingCompletion) {
	if completion == nil {
		return
	}

	ctx, lgr := log.ContextAndLoggerWithField(c.managerContext, "completion", completion)

	c.typeQuantities.Increment(completion.Identifier.Type)

	condition := &request.Condition{Revision: &completion.Identifier.Revision}

	// Validate process result, if invalid, then fail
	processResult := completion.ProcessResult

	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(&processResult); err != nil {

		// Add process result to metadata
		failedUpdateMetadata := processResult.Metadata().Clone()
		if failedUpdateMetadata == nil {
			failedUpdateMetadata = metadata.NewMetadata()
		}
		failedUpdateMetadata.Set("processResult", processResult)

		// Create failed process result
		failedUpdate := work.FailedUpdate{
			FailedError: errors.Serializable{Error: errors.New("invalid process result")},
			Metadata:    failedUpdateMetadata,
		}
		processResult = work.NewProcessResultFailed(failedUpdate)
	}

	if processResult.Result == work.ResultDelete {
		if _, err := c.workClient.Delete(context.WithoutCancel(ctx), completion.Identifier.ID, condition); err != nil {
			lgr.WithError(err).Error("Failure to delete work when processing complete")
		}
		return
	}

	wrk, err := c.workClient.Update(context.WithoutCancel(ctx), completion.Identifier.ID, condition, processResultToUpdate(processResult))
	if err != nil {
		lgr.WithError(err).Error("Failure to update work when processing complete")
	}

	lgr = lgr.WithField("work", wrk)
	switch processResult.Result {
	case work.ResultPending:
		lgr.Debug("Work state is pending")
	case work.ResultFailing:
		lgr.Warn("Work state is failing")
	case work.ResultFailed:
		lgr.Error("Work state is failed")
	case work.ResultSuccess:
		lgr.Info("Work state is success")
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
	Identifier *coordinatorProcessingIdentifier `json:"identifier,omitempty"` // Must be pointer, shared revision
}

func (c *coordinatorProcessingUpdater) ProcessingUpdate(ctx context.Context, processingUpdate work.ProcessingUpdate) (*work.Work, error) {
	condition := &request.Condition{Revision: &c.Identifier.Revision}
	update := &work.Update{
		State:            work.StateProcessing,
		ProcessingUpdate: &processingUpdate,
	}
	wrk, err := c.WorkClient.Update(context.WithoutCancel(ctx), c.Identifier.ID, condition, update)
	if err != nil {
		log.LoggerFromContext(ctx).WithError(err).Error("Failure to update work when processing")
	} else {
		c.Identifier.Revision = wrk.Revision
	}
	return wrk, err
}

type coordinatorProcessingCompletion struct {
	Identifier    *coordinatorProcessingIdentifier `json:"identifier,omitempty"` // Must be pointer, shared revision
	ProcessResult work.ProcessResult               `json:"processResult,omitempty"`
}

func processResultToUpdate(processResult work.ProcessResult) *work.Update {
	switch processResult.Result {
	case work.ResultPending:
		return &work.Update{
			State:         work.StatePending,
			PendingUpdate: processResult.PendingUpdate,
		}
	case work.ResultFailing:
		return &work.Update{
			State:         work.StateFailing,
			FailingUpdate: processResult.FailingUpdate,
		}
	case work.ResultFailed:
		return &work.Update{
			State:        work.StateFailed,
			FailedUpdate: processResult.FailedUpdate,
		}
	case work.ResultSuccess:
		return &work.Update{
			State:         work.StateSuccess,
			SuccessUpdate: processResult.SuccessUpdate,
		}
	default:
		return nil
	}
}
