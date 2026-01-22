package test

import (
	"github.com/onsi/gomega"
	gomegaGstruct "github.com/onsi/gomega/gstruct"
	gomegaTypes "github.com/onsi/gomega/types"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
)

func RandomPendingProcessResult() *work.ProcessResult {
	return work.NewProcessResultPending(work.PendingUpdate{
		ProcessingAvailableTime: test.RandomTimeAfterNow(),
		ProcessingPriority:      test.RandomInt(),
		ProcessingTimeout:       test.RandomIntFromRange(1, work.ProcessingTimeoutMaximum),
		Metadata:                metadataTest.RandomMetadataMap(),
	})
}

func RandomFailingProcessResult() *work.ProcessResult {
	return work.NewProcessResultFailing(work.FailingUpdate{
		FailingError:      *errorsTest.RandomSerializable(),
		FailingRetryCount: test.RandomIntFromRange(1, test.RandomIntMaximum()),
		FailingRetryTime:  test.RandomTimeAfterNow(),
		Metadata:          metadataTest.RandomMetadataMap(),
	})
}

func RandomFailedProcessResult() *work.ProcessResult {
	return work.NewProcessResultFailed(work.FailedUpdate{
		FailedError: *errorsTest.RandomSerializable(),
		Metadata:    metadataTest.RandomMetadataMap(),
	})
}

func RandomSuccessProcessResult() *work.ProcessResult {
	return work.NewProcessResultSuccess(work.SuccessUpdate{
		Metadata: metadataTest.RandomMetadataMap(),
	})
}

func MatchPendingProcessResult(updateMatcher gomegaTypes.GomegaMatcher) gomegaTypes.GomegaMatcher {
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"Result":        gomega.Equal(work.ResultPending),
		"PendingUpdate": gomegaGstruct.PointTo(updateMatcher),
		"FailingUpdate": gomega.BeNil(),
		"FailedUpdate":  gomega.BeNil(),
		"SuccessUpdate": gomega.BeNil(),
	}))
}

func MatchFailingProcessResult(updateMatcher gomegaTypes.GomegaMatcher) gomegaTypes.GomegaMatcher {
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"Result":        gomega.Equal(work.ResultFailing),
		"PendingUpdate": gomega.BeNil(),
		"FailingUpdate": gomegaGstruct.PointTo(updateMatcher),
		"FailedUpdate":  gomega.BeNil(),
		"SuccessUpdate": gomega.BeNil(),
	}))
}

func MatchFailedProcessResult(updateMatcher gomegaTypes.GomegaMatcher) gomegaTypes.GomegaMatcher {
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"Result":        gomega.Equal(work.ResultFailed),
		"PendingUpdate": gomega.BeNil(),
		"FailingUpdate": gomega.BeNil(),
		"FailedUpdate":  gomegaGstruct.PointTo(updateMatcher),
		"SuccessUpdate": gomega.BeNil(),
	}))
}

func MatchSuccessProcessResult(updateMatcher gomegaTypes.GomegaMatcher) gomegaTypes.GomegaMatcher {
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"Result":        gomega.Equal(work.ResultSuccess),
		"PendingUpdate": gomega.BeNil(),
		"FailingUpdate": gomega.BeNil(),
		"FailedUpdate":  gomega.BeNil(),
		"SuccessUpdate": gomegaGstruct.PointTo(updateMatcher),
	}))
}

func MatchDeleteProcessResult() gomegaTypes.GomegaMatcher {
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"Result":        gomega.Equal(work.ResultDelete),
		"PendingUpdate": gomega.BeNil(),
		"FailingUpdate": gomega.BeNil(),
		"FailedUpdate":  gomega.BeNil(),
		"SuccessUpdate": gomega.BeNil(),
	}))
}
