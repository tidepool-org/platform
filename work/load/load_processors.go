package work_load

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/work"
	"github.com/tidepool-org/platform/work/service"
)

const (
	quantity  = 1
	frequency = 5 * time.Second

	// wait and delete
	// wait and success
	// wait and run next
	// wait and run self

	// run and next

	MetadataActions        = "actions"
	MetadataActionMetadata = "actionMetadata"
	MetadataFailingData    = "failingData"

	//required metadata
	// action key - is array of objects
	// type e.g. deploy, next type of work, return this status, reschedule this

	// delay - do some work  (how long)
	// create - add another work item ()
	// return result (resturnSuccess)
	// failing & pending how many time to run before changing state
	// look in metadata to see what is needed, failing needs retry time

	// {
	// 	action:
	// 	metadata:{}
	// }
	//

	// MetadataProcessResult = "processResult"
	// MetadataSleep         = "randomSleep"
	// randomResult          = "random"

	Action = "action"

	SleepAction    = "sleep"
	ResultAction   = "result"
	CreateAction   = "createWork"
	RegisterAction = "registerProcessor"

	SleepDelay   = "delay"
	RegisterType = "type"
)

func domainName(subdomain string) string {
	return fmt.Sprintf("org.tidepool.work.test.load.%s", subdomain)
}

// TypeSleepy - will sleep for a random amount of time from 0 - `sleepMaxMillis` and then returns `ResultDelete`
var TypeSleepy = domainName("sleepy")

// TypeDopey - has to be told what to do or by default returns `ResultDelete`
var TypeDopey = domainName("dopey")

type loadProcessor struct {
	typ                   string
	quantity              int
	frequency             time.Duration
	coordinator           *service.Coordinator
	workClient            work.Client
	registerProcessorFunc func(processor work.Processor) error
}

func newLoadProcessor(typ string, quantity int, frequency time.Duration, workClient work.Client, registerProcessorFunc func(processor work.Processor) error) (*loadProcessor, error) {
	return &loadProcessor{
		typ:                   typ,
		quantity:              quantity,
		frequency:             frequency,
		workClient:            workClient,
		registerProcessorFunc: registerProcessorFunc,
	}, nil
}

func (p *loadProcessor) Type() string {
	return p.typ
}

func (p *loadProcessor) Quantity() int {
	return p.quantity
}

func (p *loadProcessor) Frequency() time.Duration {
	return p.frequency
}

func (p *loadProcessor) returnResult(name string, metadata map[string]any) *work.ProcessResult {
	switch name {
	case work.StateSuccess:
		return work.NewProcessResultSuccess(work.SuccessUpdate{
			Metadata: metadata,
		})
	case work.StatePending:

		timeout, ok := metadata["pendingTimeout"].(int)
		if !ok {
			timeout = 0
		}
		priority, ok := metadata["pendingPriority"].(int)
		if !ok {
			priority = 0
		}

		offset, ok := metadata["pendingOffsetMS"].(int)
		if !ok {
			// could be an infite loop as will kick asap
			offset = 0
		}

		return work.NewProcessResultPending(work.PendingUpdate{
			ProcessingTimeout:       timeout,
			Metadata:                metadata,
			ProcessingPriority:      priority,
			ProcessingAvailableTime: time.Now().Add(time.Millisecond * time.Duration(offset)),
		})
	case work.StateFailed:
		msg, ok := metadata["failedMessage"].(string)
		if !ok {
			msg = "failure from return result"
		}
		return work.NewProcessResultFailed(work.FailedUpdate{
			FailedError: *errors.NewSerializable(errors.New(msg)),
			Metadata:    metadata,
		})
	case work.StateFailing:
		msg, ok := metadata["failingMessage"].(string)
		if !ok {
			msg = "failing from return result"
		}
		count, ok := metadata["failingRetryCount"].(int)
		if !ok {
			count = 3
		}
		offset, ok := metadata["failingOffsetMS"].(int)
		if !ok {
			offset = 3
		}
		return work.NewProcessResultFailing(work.FailingUpdate{
			FailingError:      *errors.NewSerializable(errors.New(msg)),
			FailingRetryCount: count,
			FailingRetryTime:  time.Now().Add(time.Millisecond * time.Duration(offset)),
			Metadata:          metadata,
		})

	default:
		return work.NewProcessResultFailed(work.FailedUpdate{
			FailedError: *errors.NewSerializable(fmt.Errorf("unknown result type %s ", name)),
			Metadata:    metadata,
		})
	}
}

func (p *loadProcessor) performAction(ctx context.Context, name string, metadata map[string]any) *work.ProcessResult {

	switch name {

	case SleepAction:
		delayMillisecond, ok := metadata[SleepDelay].(int)
		if !ok {
			delayMillisecond = rand.IntN(1000)
		}
		time.Sleep(time.Duration(delayMillisecond) * time.Millisecond)
		return nil
	case ResultAction:
		result, ok := metadata["result"].(string)
		if !ok {
			return p.returnResult(work.ResultFailed, metadata)
		}
		return p.returnResult(result, metadata)
	case CreateAction:
		if p.coordinator == nil {
			p.returnResult(work.ResultFailed, metadata)
		}
		create, ok := metadata["create"].(work.Create)
		if !ok {
			p.returnResult(work.ResultFailed, metadata)
		}
		_, err := p.workClient.Create(ctx, &create)
		if err != nil {
			p.returnResult(work.ResultFailed, metadata)
		}
		return nil
	case RegisterAction:
		//TODO: register a new processor
		registerType, ok := metadata[RegisterType].(string)
		if !ok {
			p.returnResult(work.ResultFailed, metadata)
		}

		newProcessor, err := newLoadProcessor(registerType, quantity, frequency, p.workClient, p.registerProcessorFunc)
		if err != nil {
			p.returnResult(work.ResultFailed, metadata)
		}

		if err := p.registerProcessorFunc(newProcessor); err != nil {
			p.returnResult(work.ResultFailed, metadata)
		}
	}

	return p.returnResult(work.ResultFailed, metadata)
}

func (p *loadProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {

	if wrk.Type != TypeSleepy && wrk.Type != TypeDopey {
		p.returnResult(work.ResultFailed, wrk.Metadata)
	}

	actions, ok := wrk.Metadata[MetadataActions].([]map[string]any)
	if !ok {
		return *p.returnResult(work.ResultSuccess, wrk.Metadata)
	}

	for _, actionData := range actions {
		name, ok := actionData[Action].(string)
		if ok {
			if result := p.performAction(ctx, name, actionData); result != nil {
				return *result
			}
		}
	}
	return *p.returnResult(work.ResultDelete, wrk.Metadata)
}

type LoadItem struct {
	OffsetMilliseconds int64        `json:"offsetMilliseconds"`
	Create             *work.Create `json:"create"`
}

func NewLoadProcessors(workClient work.Client, registerProcessorFunc func(processor work.Processor) error) ([]work.Processor, error) {
	var processors []work.Processor
	if sleepyProcessor, err := newLoadProcessor(TypeSleepy, quantity, frequency, workClient, registerProcessorFunc); err != nil {
		return nil, err
	} else {
		processors = append(processors, sleepyProcessor)
	}

	if dopeyProcessor, err := newLoadProcessor(TypeDopey, quantity, frequency, workClient, registerProcessorFunc); err != nil {
		return nil, err
	} else {
		processors = append(processors, dopeyProcessor)
	}

	return processors, nil
}
