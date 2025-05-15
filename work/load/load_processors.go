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
	ResultAction   = "result"
	CreateAction   = "createWork"
	RegisterAction = "registerProcessor"

	SleepDelay   = "delay"
	RegisterType = "type"
	CreateType   = "type"
)

const domainPrefix = "org.tidepool.work.test.load"

func DomainName(subdomain string) string {
	return fmt.Sprintf("%s.%s", domainPrefix, subdomain)
}

var TypeSleepy = DomainName("sleepy")
var TypeDopey = DomainName("dopey")

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

	switch name {

	case SleepAction:
		delayMillisecond, ok := metadata[SleepDelay].(float64)
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
		parser := structureParser.NewObject(logNull.NewLogger(), &createObj)

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

	actions := Actions{}
	if err := json.Unmarshal(jsonData, &actions); err != nil {
		p.returnResult(work.ResultFailed, wrk.Metadata, err)
	}

	for _, actionData := range actions {
		name, ok := actionData[ActionKey].(string)
		if ok {
			if result := p.performAction(ctx, wrk, name, actionData); result != nil {
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

type Actions []Action

type Action map[string]any

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
