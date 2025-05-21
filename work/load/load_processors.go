package work_load

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/tidepool-org/platform/errors"
	logNull "github.com/tidepool-org/platform/log/null"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	"github.com/tidepool-org/platform/work"
	"github.com/tidepool-org/platform/work/service"
)

const (
	quantity  = 1
	frequency = 5 * time.Second

	MetadataActions = "actions"

	ActionKey      = "action"
	SleepAction    = "sleep"
	FailureAction  = "failure"
	ResultAction   = "result"
	CreateAction   = "createWork"
	RegisterAction = "registerProcessor"

	FailureOffsetMS   = "failureOffsetMilliseconds"
	FailureDurationMS = "failureDurationMilliseconds"
	FailingOffsetMS   = "failingOffsetMilliseconds"
	PendingOffsetMS   = "pendingOffsetMilliseconds"
	CreateCount       = "createWorkCount"
	SleepDelayMS      = "delayMilliseconds"
	RegisterType      = "type"
	CreateType        = "type"
)

const domainPrefix = "org.tidepool.work.test.load"

func DomainName(subdomain string) string {
	return fmt.Sprintf("%s.%s", domainPrefix, subdomain)
}

var TypeSleepy = DomainName("sleepy")
var TypeDopey = DomainName("dopey")

func AllowedActions() []string {
	return []string{
		SleepAction,
		FailureAction,
		ResultAction,
		CreateAction,
		RegisterAction,
	}
}

type loadProcessor struct {
	typ                   string
	quantity              int
	failureStart          time.Time
	failureEnd            time.Time
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
		failureStart:          time.Time{},
		failureEnd:            time.Time{},
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

func (p *loadProcessor) returnResult(name string, metadata map[string]any, err error) *work.ProcessResult {
	switch name {
	case work.ResultSuccess:
		return work.NewProcessResultSuccess(work.SuccessUpdate{
			Metadata: metadata,
		})
	case work.ResultPending:

		timeout, ok := metadata["pendingTimeout"].(int)
		if !ok {
			timeout = 0
		}
		priority, ok := metadata["pendingPriority"].(int)
		if !ok {
			priority = 0
		}

		offsetMS, ok := metadata[PendingOffsetMS].(int)
		if !ok {
			offsetMS = 0
		}

		return work.NewProcessResultPending(work.PendingUpdate{
			ProcessingTimeout:       timeout,
			Metadata:                metadata,
			ProcessingPriority:      priority,
			ProcessingAvailableTime: time.Now().Add(time.Millisecond * time.Duration(offsetMS)),
		})
	case work.ResultFailed:
		if err == nil {
			msg, ok := metadata["failedMessage"].(string)
			if !ok {
				msg = "failure from return result"
			}
			err = errors.New(msg)
		}

		return work.NewProcessResultFailed(work.FailedUpdate{
			FailedError: *errors.NewSerializable(err),
			Metadata:    metadata,
		})
	case work.ResultFailing:
		if err == nil {
			msg, ok := metadata["failingMessage"].(string)
			if !ok {
				msg = "failing from return result"
			}
			err = errors.New(msg)
		}
		count, ok := metadata["failingRetryCount"].(int)
		if !ok {
			count = 3
		}
		retryOffset, ok := metadata["failingOffsetMS"].(int)
		if !ok {
			retryOffset = 30 * 1000
		}
		return work.NewProcessResultFailing(work.FailingUpdate{
			FailingError:      *errors.NewSerializable(err),
			FailingRetryCount: count,
			FailingRetryTime:  time.Now().Add(time.Millisecond * time.Duration(retryOffset)),
			Metadata:          metadata,
		})
	case work.ResultDelete:
		return work.NewProcessResultDelete()
	default:
		return work.NewProcessResultFailed(work.FailedUpdate{
			FailedError: *errors.NewSerializable(fmt.Errorf("unknown result type %s", name)),
			Metadata:    metadata,
		})
	}
}

func (p *loadProcessor) performAction(ctx context.Context, wrk *work.Work, name string, metadata map[string]any) *work.ProcessResult {

	if !p.failureStart.IsZero() {
		now := time.Now()
		if p.failureStart.Before(now) && p.failureEnd.After(now) {
			return p.returnResult(
				work.ResultFailing,
				metadata,
				fmt.Errorf("system failure from %s until %s", p.failureStart.Format(time.RFC3339), p.failureEnd.Format(time.RFC3339)),
			)
		}
	}

	switch name {
	case FailureAction:
		offsetMillisecond, ok := metadata[FailureOffsetMS].(float64)
		if !ok {
			offsetMillisecond = float64(rand.IntN(1000))
		}

		durationMillisecond, ok := metadata[FailureDurationMS].(float64)
		if !ok {
			durationMillisecond = float64(rand.IntN(1000))
		}
		p.failureStart = wrk.ProcessingTime.Add(time.Millisecond * time.Duration(offsetMillisecond))
		p.failureEnd = p.failureStart.Add(time.Millisecond * time.Duration(durationMillisecond))
		return nil
	case SleepAction:
		delayMillisecond, ok := metadata[SleepDelayMS].(float64)
		if !ok {
			delayMillisecond = float64(rand.IntN(1000))
		}
		time.Sleep(time.Duration(delayMillisecond) * time.Millisecond)
		return nil
	case ResultAction:
		result, ok := metadata["result"].(string)
		if !ok {
			return p.returnResult(work.ResultFailed, metadata, errors.New("result not found"))
		}
		return p.returnResult(result, metadata, nil)
	case CreateAction:
		if p.coordinator == nil {
			p.returnResult(work.ResultFailed, metadata, errors.New("coordinator not set"))
		}
		createObj := metadata["create"].(map[string]any)
		createCount, ok := metadata[CreateCount].(float64)
		if !ok {
			createCount = 1
		}

		parser := structureParser.NewObject(logNull.NewLogger(), &createObj)

		for range int(createCount) {
			create := work.ParseCreate(parser)
			if create.GroupID == nil {
				create.GroupID = wrk.GroupID
			}
			if create.SerialID == nil {
				create.SerialID = wrk.SerialID
			}
			_, err := p.workClient.Create(ctx, create)
			if err != nil {
				p.returnResult(work.ResultFailed, metadata, err)
			}
		}

		return nil
	case RegisterAction:
		//TODO: register a new processor
		registerType, ok := metadata[RegisterType].(string)
		if !ok {
			p.returnResult(work.ResultFailed, metadata, fmt.Errorf("%s has invalid type", RegisterAction))
		}
		newProcessor, err := newLoadProcessor(registerType, quantity, frequency, p.workClient, p.registerProcessorFunc)
		if err != nil {
			p.returnResult(work.ResultFailed, metadata, err)
		}

		if err := p.registerProcessorFunc(newProcessor); err != nil {
			p.returnResult(work.ResultFailed, metadata, err)
		}
		return nil
	}

	return p.returnResult(work.ResultFailed, metadata, fmt.Errorf("unknown action name %s", name))
}

func (p *loadProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	if strings.Contains(wrk.Type, domainPrefix) {
		p.returnResult(work.ResultFailed, wrk.Metadata, errors.New("invalid work type"))
	}

	metadataActions := wrk.Metadata[MetadataActions]

	jsonData, err := json.Marshal(metadataActions)
	if err != nil {
		p.returnResult(work.ResultFailed, wrk.Metadata, err)
	}

	processActions := ProcessActions{}
	if err := json.Unmarshal(jsonData, &processActions); err != nil {
		p.returnResult(work.ResultFailed, wrk.Metadata, err)
	}

	for _, processAction := range processActions {
		name, ok := processAction[ActionKey].(string)
		if ok {
			if result := p.performAction(ctx, wrk, name, processAction); result != nil {
				return *result
			}
		}
	}
	return *p.returnResult(work.ResultSuccess, wrk.Metadata, nil)
}

type LoadItem struct {
	OffsetMilliseconds int64        `json:"offsetMilliseconds"`
	Create             *work.Create `json:"create"`
}

type ProcessActions []ProcessAction

type ProcessAction map[string]any

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
