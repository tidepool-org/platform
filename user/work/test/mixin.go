package test

import (
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	userWork "github.com/tidepool-org/platform/user/work"
)

func RandomMetadata(options ...test.Option) *userWork.Metadata {
	return &userWork.Metadata{
		UserID: test.RandomOptional(userTest.RandomUserID, options...),
	}
}

func CloneMetadata(datum *userWork.Metadata) *userWork.Metadata {
	if datum == nil {
		return nil
	}
	return &userWork.Metadata{
		UserID: pointer.Clone(datum.UserID),
	}
}

func NewObjectFromMetadata(datum *userWork.Metadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.UserID != nil {
		object[userWork.MetadataKeyUserID] = test.NewObjectFromString(*datum.UserID, objectFormat)
	}
	return object
}
