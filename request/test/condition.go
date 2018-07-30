package test

import (
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
)

func RandomRevision() int {
	return test.RandomIntFromRange(0, test.RandomIntMaximum())
}

func RandomCondition() *request.Condition {
	datum := &request.Condition{}
	datum.Revision = pointer.FromInt(RandomRevision())
	return datum
}

func NewObjectFromCondition(datum *request.Condition, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Revision != nil {
		object["revision"] = test.NewObjectFromInt(*datum.Revision, objectFormat)
	}
	return object
}
