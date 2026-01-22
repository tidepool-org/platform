package test

import (
	time "time"

	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
)

func RandomState() string {
	return test.RandomStringFromArray(work.States())
}

func RandomID() string {
	return bsonPrimitive.NewObjectID().String()
}

func RandomType() string {
	return test.RandomString()
}

func RandomGroupID() string {
	return test.RandomString()
}

func RandomSerialID() string {
	return test.RandomString()
}

func RandomDeduplicationID() string {
	return test.RandomString()
}

func RandomWork() *work.Work {
	return RandomWorkWithState(RandomState())
}

func RandomWorkWithState(state string) *work.Work {
	now := time.Now()
	createdTime := test.RandomTimeBefore(now)

	datum := &work.Work{}
	datum.ID = RandomID()
	datum.Type = RandomType()
	datum.GroupID = pointer.FromString(RandomGroupID())
	datum.DeduplicationID = pointer.FromString(RandomDeduplicationID())
	datum.SerialID = pointer.FromString(RandomSerialID())
	datum.ProcessingPriority = test.RandomInt()
	datum.ProcessingTimeout = test.RandomIntFromRange(1, 15*60)
	datum.Metadata = metadataTest.RandomMetadataMap()
	datum.State = state
	datum.CreatedTime = createdTime
	datum.ModifiedTime = pointer.FromTime(now)
	datum.Revision = test.RandomIntFromRange(0, test.RandomIntMaximum())
	if state == work.StatePending {
		datum.PendingTime = now
		datum.ProcessingAvailableTime = test.RandomTimeAfter(now)
	} else {
		datum.PendingTime = test.RandomTimeFromRange(createdTime, now)
		datum.ProcessingAvailableTime = test.RandomTimeFromRange(datum.PendingTime, now)
		if state == work.StateProcessing {
			datum.ProcessingTime = pointer.FromTime(now)
			datum.ProcessingTimeoutTime = pointer.FromTime(test.RandomTimeAfter(now))
		} else {
			datum.ProcessingTime = pointer.FromTime(test.RandomTimeFromRange(datum.ProcessingAvailableTime, now))
			datum.ProcessingDuration = pointer.FromFloat64(now.Sub(*datum.ProcessingTime).Seconds())
			if state == work.StateFailing {
				datum.FailingTime = pointer.FromTime(now)
				datum.FailingError = errors.NewSerializable(errorsTest.RandomError())
				datum.FailingRetryCount = pointer.FromInt(test.RandomIntFromRange(0, 10))
				datum.FailingRetryTime = pointer.FromTime(test.RandomTimeAfter(now))
			} else if state == work.StateFailed {
				datum.FailedTime = pointer.FromTime(now)
				datum.FailedError = errors.NewSerializable(errorsTest.RandomError())
			} else if state == work.StateSuccess {
				datum.SuccessTime = pointer.FromTime(now)
			}
		}
	}
	return datum
}

func CloneWork(datum *work.Work) *work.Work {
	if datum == nil {
		return nil
	}
	clone := &work.Work{}
	clone.ID = datum.ID
	clone.Type = datum.Type
	clone.GroupID = pointer.CloneString(datum.GroupID)
	clone.DeduplicationID = pointer.CloneString(datum.DeduplicationID)
	clone.SerialID = pointer.CloneString(datum.SerialID)
	clone.ProcessingAvailableTime = datum.ProcessingAvailableTime
	clone.ProcessingPriority = datum.ProcessingPriority
	clone.ProcessingTimeout = datum.ProcessingTimeout
	clone.Metadata = metadataTest.CloneMetadataMap(datum.Metadata)
	clone.PendingTime = datum.PendingTime
	clone.ProcessingTime = pointer.CloneTime(datum.ProcessingTime)
	clone.ProcessingTimeoutTime = pointer.CloneTime(datum.ProcessingTimeoutTime)
	clone.ProcessingDuration = pointer.CloneFloat64(datum.ProcessingDuration)
	clone.FailingTime = pointer.CloneTime(datum.FailingTime)
	clone.FailingError = errorsTest.CloneSerializable(datum.FailingError)
	clone.FailingRetryCount = pointer.CloneInt(datum.FailingRetryCount)
	clone.FailingRetryTime = pointer.CloneTime(datum.FailingRetryTime)
	clone.FailedTime = pointer.CloneTime(datum.FailedTime)
	clone.FailedError = errorsTest.CloneSerializable(datum.FailedError)
	clone.SuccessTime = pointer.CloneTime(datum.SuccessTime)
	clone.State = datum.State
	clone.CreatedTime = datum.CreatedTime
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	clone.Revision = datum.Revision
	return clone
}

func NewObjectFromWork(datum *work.Work, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	object["id"] = test.NewObjectFromString(datum.ID, objectFormat)
	object["type"] = test.NewObjectFromString(datum.Type, objectFormat)
	if datum.GroupID != nil {
		object["groupId"] = test.NewObjectFromString(*datum.GroupID, objectFormat)
	}
	if datum.DeduplicationID != nil {
		object["deduplicationId"] = test.NewObjectFromString(*datum.DeduplicationID, objectFormat)
	}
	if datum.SerialID != nil {
		object["serialId"] = test.NewObjectFromString(*datum.SerialID, objectFormat)
	}
	object["processingAvailableTime"] = test.NewObjectFromTime(datum.ProcessingAvailableTime, objectFormat)
	object["processingPriority"] = test.NewObjectFromInt(datum.ProcessingPriority, objectFormat)
	object["processingTimeout"] = test.NewObjectFromInt(datum.ProcessingTimeout, objectFormat)
	object["metadata"] = metadataTest.NewObjectFromMetadataMap(datum.Metadata, objectFormat)
	object["pendingTime"] = test.NewObjectFromTime(datum.PendingTime, objectFormat)
	if datum.ProcessingTime != nil {
		object["processingTime"] = test.NewObjectFromTime(*datum.ProcessingTime, objectFormat)
	}
	if datum.ProcessingTimeoutTime != nil {
		object["processingTimeoutTime"] = test.NewObjectFromTime(*datum.ProcessingTimeoutTime, objectFormat)
	}
	if datum.ProcessingDuration != nil {
		object["processingDuration"] = test.NewObjectFromFloat64(*datum.ProcessingDuration, objectFormat)
	}
	if datum.FailingTime != nil {
		object["failingTime"] = test.NewObjectFromTime(*datum.FailingTime, objectFormat)
	}
	if datum.FailingError != nil {
		object["failingError"] = errorsTest.NewObjectFromSerializable(datum.FailingError, objectFormat)
	}
	if datum.FailingRetryCount != nil {
		object["failingRetryCount"] = test.NewObjectFromInt(*datum.FailingRetryCount, objectFormat)
	}
	if datum.FailingRetryTime != nil {
		object["failingRetryTime"] = test.NewObjectFromTime(*datum.FailingRetryTime, objectFormat)
	}
	if datum.FailedTime != nil {
		object["failedTime"] = test.NewObjectFromTime(*datum.FailedTime, objectFormat)
	}
	if datum.FailedError != nil {
		object["failedError"] = errorsTest.NewObjectFromSerializable(datum.FailedError, objectFormat)
	}
	if datum.SuccessTime != nil {
		object["successTime"] = test.NewObjectFromTime(*datum.SuccessTime, objectFormat)
	}
	object["state"] = test.NewObjectFromString(datum.State, objectFormat)
	object["createdTime"] = test.NewObjectFromTime(datum.CreatedTime, objectFormat)
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	object["revision"] = test.NewObjectFromInt(datum.Revision, objectFormat)

	return object
}
