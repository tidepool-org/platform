package workTestLoad

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
)

const (
	quantity  = 1
	frequency = 5 * time.Second

	MetadataProcessResult = "processResult"
	MetadataSleep         = "randomSleep"
	randomResult          = "random"
)

func domainName(subdomain string) string {
	return fmt.Sprintf("org.tidepool.work.test.load.%s", subdomain)
}

var TypeSleepy = domainName("sleepy")
var TypeDopey = domainName("dopey")

type loadProcessor struct {
	typ       string
	quantity  int
	frequency time.Duration
}

func newLoadProcessor(typ string, quantity int, frequency time.Duration) (*loadProcessor, error) {
	return &loadProcessor{
		typ:       typ,
		quantity:  quantity,
		frequency: frequency,
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

func (p *loadProcessor) getProcessResult(wrk *work.Work, result any, errMsg *string) *work.ProcessResult {

	if result.(string) == randomResult {
		indexnum := rand.IntN(len(work.Results()) - 1)
		result = work.Results()[indexnum]
	}

	switch result.(string) {
	case work.ResultFailing:

		msg := "failing from load test"
		if errMsg != nil {
			msg = *errMsg
		}

		return work.NewProcessResultFailing(work.FailingUpdate{
			FailingError:      *errors.NewSerializable(errors.New(msg)),
			FailingRetryCount: 2,
			Metadata:          wrk.Metadata,
		})
	case work.ResultFailed:
		return work.NewProcessResultFailed(work.FailedUpdate{
			FailedError: *errors.NewSerializable(errors.New("failed from load test")),
			Metadata:    wrk.Metadata,
		})
	case work.ResultSuccess:
		return work.NewProcessResultSuccess(work.SuccessUpdate{
			Metadata: wrk.Metadata,
		})
	case work.ResultPending:
		return work.NewProcessResultPending(work.PendingUpdate{
			Metadata: wrk.Metadata,
		})
	default:
		return work.NewProcessResultDelete()
	}
}

func (p *loadProcessor) getResult(wrk *work.Work) string {
	if result := wrk.Metadata[MetadataProcessResult]; result != nil {
		if resultStr, ok := result.(string); ok {
			if slices.Contains(work.Results(), resultStr) {
				return resultStr
			}
		}
	}
	return randomResult
}

func (p *loadProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	switch wrk.Type {
	case TypeSleepy, TypeDopey:
		if sleep := wrk.Metadata[MetadataSleep]; sleep != nil {
			waitMillis := rand.IntN(500)
			time.Sleep(time.Duration(waitMillis) * time.Millisecond)
		}
		return *p.getProcessResult(wrk, p.getResult(wrk), nil)
	default:
		return *p.getProcessResult(wrk, work.ResultFailed, pointer.FromString(fmt.Sprintf("[%s] not an expected work type", wrk.Type)))
	}
}

type LoadItem struct {
	OffsetMilliseconds int64        `json:"offsetMilliseconds"`
	Create             *work.Create `json:"create"`
}

type ActionItem struct {
	Name     string         `json:"actionName"`
	Metadata map[string]any `json:"actionMetadata"`
	Result   *ActionItem    `json:"result"`
}

func NewLoadProcessors() ([]work.Processor, error) {
	var processors []work.Processor
	if sleepyProcessor, err := newLoadProcessor(TypeSleepy, quantity, frequency); err != nil {
		return nil, err
	} else {
		processors = append(processors, sleepyProcessor)
	}

	if dopeyProcessor, err := newLoadProcessor(TypeDopey, quantity, frequency); err != nil {
		return nil, err
	} else {
		processors = append(processors, dopeyProcessor)
	}

	return processors, nil
}
