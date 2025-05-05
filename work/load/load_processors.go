package load

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	work "github.com/tidepool-org/platform/work"
)

const (
	quantity                     = 1
	frequency                    = 5 * time.Second
	processingAvailableTimeDelay = 5 * time.Second
	processingTimeout            = 5 * 60

	groupPrefix = "org.tidepool.work.test.load"

	metadataProcessResult = "processResult"
	metadataDataSessionID = "sessionId"
)

func workSerialIDFromSessionID(wrkType string, sessionID string) string {
	return fmt.Sprintf("%s:%s", domainName(wrkType), sessionID)
}

func domainName(subdomain string) string {
	return fmt.Sprintf("%s.%s", groupPrefix, subdomain)
}

func workGroupIDFromSessionID(sessionID string) string {
	return fmt.Sprintf("%s:%s", groupPrefix, sessionID)
}

// TypeSleepy - will sleep for a random amount of time from 0 - `sleepMaxMillis` and then returns `ResultDelete`
var TypeSleepy = domainName("sleepy")

// TypeDopey - has to be told what to do or by default returns `ResultDelete`
var TypeDopey = domainName("dopey")

type loadProcessor struct {
	pType string
}

func newLoadProcessor(pType string) (*loadProcessor, error) {
	if pType != TypeSleepy && pType != TypeDopey {
		return nil, errors.Newf("type %s is invalid. Must be one of [%s,%s]", pType, TypeSleepy, TypeDopey)
	}
	return &loadProcessor{
		pType: pType,
	}, nil
}

func (p *loadProcessor) Type() string {
	return p.pType
}

func (p *loadProcessor) Quantity() int {
	return quantity
}

func (p *loadProcessor) Frequency() time.Duration {
	return frequency
}

func (p *loadProcessor) getProcessResult(wrk *work.Work, result any, errMsg *string) *work.ProcessResult {
	switch result.(string) {
	case work.ResultFailing:

		msg := "failing from load test"
		if errMsg != nil {
			msg = *errMsg
		}

		return work.NewProcessResultFailing(work.FailingUpdate{
			FailingError: *errors.NewSerializable(errors.New(msg)),
			Metadata:     wrk.Metadata,
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
	default:
		return work.NewProcessResultDelete()
	}
}

func (p *loadProcessor) chooseDopeyResult(wrk *work.Work) string {
	possibleWorkResults := []string{
		work.ResultDelete,
		work.ResultFailing,
		work.ResultFailed,
		work.ResultSuccess,
	}
	if result := wrk.Metadata[metadataProcessResult]; result != nil {
		if resultStr, ok := result.(string); ok {
			if resultStr == "random" {
				var indexnum int = rand.IntN(len(possibleWorkResults) - 1)
				return possibleWorkResults[indexnum]
			}
			if slices.Contains(possibleWorkResults, resultStr) {
				return resultStr
			}
		}
	}
	return work.ResultDelete
}

func (p *loadProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	switch wrk.Type {
	case TypeSleepy:
		waitMillis := rand.IntN(500)
		time.Sleep(time.Duration(waitMillis) * time.Millisecond)
		return *p.getProcessResult(wrk, work.ResultSuccess, nil)
	case TypeDopey:
		return *p.getProcessResult(wrk, p.chooseDopeyResult(wrk), nil)
	default:
		return *p.getProcessResult(wrk, work.ResultFailed, pointer.FromString(fmt.Sprintf("[%s] not an expected work type", wrk.Type)))
	}
}

type LoadItem struct {
	LoadProcessID          string       `json:"loadProcessId"`
	SessionID              string       `json:"sessionId"`
	SecondsOffsetFromStart int64        `json:"secondsOffsetFromStart"`
	Create                 *work.Create `json:"create"`
}

func NewLoadWorkCreate(sessionID string, loadProcessID string, create *work.Create) (*work.Create, error) {
	if sessionID == "" {
		return nil, errors.New("session id is missing")
	}
	if loadProcessID == "" {
		return nil, errors.New("load process id is missing")
	}
	if create.Type != TypeSleepy && create.Type != TypeDopey {
		return nil, fmt.Errorf("invalid work type %s", create.Type)
	}

	availableTime := time.Now().Add(processingAvailableTimeDelay)
	if !create.ProcessingAvailableTime.IsZero() {
		availableTime = create.ProcessingAvailableTime.Add(processingAvailableTimeDelay)
	}
	timeout := processingTimeout
	if create.ProcessingTimeout != 0 {
		timeout = create.ProcessingTimeout
	}

	metadata := map[string]any{
		metadataDataSessionID: sessionID,
	}

	if create.Metadata != nil && create.Metadata[metadataProcessResult] != nil {
		metadata[metadataProcessResult] = create.Metadata[metadataProcessResult]
	}

	return &work.Create{
		Type:                    create.Type,
		GroupID:                 pointer.FromString(workGroupIDFromSessionID(sessionID)),
		DeduplicationID:         pointer.FromString(fmt.Sprintf("%s:%s", sessionID, loadProcessID)),
		SerialID:                pointer.FromString(workSerialIDFromSessionID(create.Type, sessionID)),
		ProcessingAvailableTime: availableTime,
		ProcessingTimeout:       timeout,
		Metadata:                metadata,
	}, nil
}

func NewLoadProcessors() ([]work.Processor, error) {
	var processors []work.Processor
	if sleepyProcessor, err := newLoadProcessor(TypeSleepy); err != nil {
		return nil, err
	} else {
		processors = append(processors, sleepyProcessor)
	}

	if dopeyProcessor, err := newLoadProcessor(TypeDopey); err != nil {
		return nil, err
	} else {
		processors = append(processors, dopeyProcessor)
	}

	return processors, nil
}
