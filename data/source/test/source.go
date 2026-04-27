package test

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/onsi/gomega"
	gomegaGstruct "github.com/onsi/gomega/gstruct"
	gomegaTypes "github.com/onsi/gomega/types"

	authTest "github.com/tidepool-org/platform/auth/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	requestTest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomDataSourceID() string {
	return dataSource.NewID()
}

func RandomState() string {
	return test.RandomStringFromArray(dataSource.States())
}

func RandomStates() []string {
	return test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(dataSource.States()), dataSource.States())
}

func RandomDeviceID() string {
	return test.RandomString()
}

func RandomDeviceHash() string {
	md5Sum := md5.Sum([]byte(test.RandomString()))
	return hex.EncodeToString(md5Sum[:])
}

func RandomDeviceHashMap() map[string]any {
	deviceHashMap := map[string]any{}
	for range test.RandomIntFromRange(1, 3) {
		deviceHashMap[RandomDeviceID()] = RandomDeviceHash()
	}
	return deviceHashMap
}

func RandomFilter(options ...test.Option) *dataSource.Filter {
	return &dataSource.Filter{
		ProviderType:       test.RandomOptional(authTest.RandomProviderType, options...),
		ProviderName:       test.RandomOptional(authTest.RandomProviderName, options...),
		ProviderExternalID: test.RandomOptional(authTest.RandomProviderExternalID, options...),
		State:              test.RandomOptional(RandomState, options...),
	}
}

func CloneFilter(datum *dataSource.Filter) *dataSource.Filter {
	if datum == nil {
		return nil
	}
	return &dataSource.Filter{
		ProviderType:       pointer.CloneString(datum.ProviderType),
		ProviderName:       pointer.CloneString(datum.ProviderName),
		ProviderExternalID: pointer.CloneString(datum.ProviderExternalID),
		State:              pointer.CloneString(datum.State),
	}
}

func NewObjectFromFilter(datum *dataSource.Filter, objectFormat test.ObjectFormat) map[string]any {
	object := map[string]any{}
	if datum.ProviderType != nil {
		object["providerType"] = test.NewObjectFromString(*datum.ProviderType, objectFormat)
	}
	if datum.ProviderName != nil {
		object["providerName"] = test.NewObjectFromString(*datum.ProviderName, objectFormat)
	}
	if datum.ProviderExternalID != nil {
		object["providerExternalId"] = test.NewObjectFromString(*datum.ProviderExternalID, objectFormat)
	}
	if datum.State != nil {
		object["state"] = test.NewObjectFromString(*datum.State, objectFormat)
	}
	return object
}

func RandomCreate(options ...test.Option) *dataSource.Create {
	return &dataSource.Create{
		ProviderType:       authTest.RandomProviderType(),
		ProviderName:       authTest.RandomProviderName(),
		ProviderExternalID: test.RandomOptional(authTest.RandomProviderExternalID, options...),
		Metadata:           metadataTest.RandomMetadataMap(),
	}
}

func CloneCreate(datum *dataSource.Create) *dataSource.Create {
	if datum == nil {
		return nil
	}
	return &dataSource.Create{
		ProviderType:       datum.ProviderType,
		ProviderName:       datum.ProviderName,
		ProviderExternalID: pointer.CloneString(datum.ProviderExternalID),
		Metadata:           metadataTest.CloneMetadataMap(datum.Metadata),
	}
}

func NewObjectFromCreate(datum *dataSource.Create, objectFormat test.ObjectFormat) map[string]any {
	object := map[string]any{}
	object["providerType"] = test.NewObjectFromString(datum.ProviderType, objectFormat)
	object["providerName"] = test.NewObjectFromString(datum.ProviderName, objectFormat)
	if datum.ProviderExternalID != nil {
		object["providerExternalId"] = test.NewObjectFromString(*datum.ProviderExternalID, objectFormat)
	}
	if len(datum.Metadata) > 0 {
		object["metadata"] = metadataTest.NewObjectFromMetadataMap(datum.Metadata, objectFormat)
	}
	return object
}

func RandomUpdate(options ...test.Option) *dataSource.Update {
	var state *string
	var earliestDataTime *time.Time
	var latestDataTime *time.Time
	var lastImportTime *time.Time

	state = test.RandomOptional(RandomState, options...)
	now := time.Now()
	switch test.RandomIntFromRange(0, 4) {
	case 0:
	case 1:
		lastImportTime = pointer.FromTime(test.RandomTimeBeforeNow())
	case 2:
		earliestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
		lastImportTime = pointer.FromTime(test.RandomTimeFromRange(*earliestDataTime, now))
	case 3:
		latestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
		lastImportTime = pointer.FromTime(test.RandomTimeFromRange(*latestDataTime, now))
	case 4:
		earliestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
		latestDataTime = pointer.FromTime(test.RandomTimeFromRange(*earliestDataTime, now))
		lastImportTime = pointer.FromTime(test.RandomTimeFromRange(*latestDataTime, now))
	}

	return &dataSource.Update{
		ProviderExternalID: test.RandomOptional(authTest.RandomProviderExternalID, options...),
		ProviderSessionID:  test.Conditional(authTest.RandomProviderSessionID, state != nil && *state == dataSource.StateConnected),
		State:              state,
		Metadata:           test.RandomOptional(metadataTest.RandomMetadataMap, options...),
		Error:              test.RandomOptionalPointer(errorsTest.RandomSerializable, options...),
		DataSetID:          test.RandomOptional(dataTest.RandomDataSetID, options...),
		EarliestDataTime:   earliestDataTime,
		LatestDataTime:     latestDataTime,
		LastImportTime:     lastImportTime,
	}
}

func CloneUpdate(datum *dataSource.Update) *dataSource.Update {
	if datum == nil {
		return nil
	}
	return &dataSource.Update{
		ProviderExternalID: pointer.CloneString(datum.ProviderExternalID),
		ProviderSessionID:  pointer.CloneString(datum.ProviderSessionID),
		State:              pointer.CloneString(datum.State),
		Metadata:           metadataTest.CloneMetadataMapPointer(datum.Metadata),
		Error:              errorsTest.CloneSerializable(datum.Error),
		DataSetID:          pointer.CloneString(datum.DataSetID),
		EarliestDataTime:   pointer.CloneTime(datum.EarliestDataTime),
		LatestDataTime:     pointer.CloneTime(datum.LatestDataTime),
		LastImportTime:     pointer.CloneTime(datum.LastImportTime),
	}
}

func NewObjectFromUpdate(datum *dataSource.Update, objectFormat test.ObjectFormat) map[string]any {
	object := map[string]any{}
	if datum.ProviderExternalID != nil {
		object["providerExternalId"] = test.NewObjectFromString(*datum.ProviderExternalID, objectFormat)
	}
	if datum.ProviderSessionID != nil {
		object["providerSessionId"] = test.NewObjectFromString(*datum.ProviderSessionID, objectFormat)
	}
	if datum.State != nil {
		object["state"] = test.NewObjectFromString(*datum.State, objectFormat)
	}
	if datum.Metadata != nil {
		object["metadata"] = metadataTest.NewObjectFromMetadataMap(*datum.Metadata, objectFormat)
	}
	if datum.Error != nil {
		object["error"] = errorsTest.NewObjectFromSerializable(datum.Error, objectFormat)
	}
	if datum.DataSetID != nil {
		object["dataSetId"] = test.NewObjectFromString(*datum.DataSetID, objectFormat)
	}
	if datum.EarliestDataTime != nil {
		object["earliestDataTime"] = test.NewObjectFromTime(*datum.EarliestDataTime, objectFormat)
	}
	if datum.LatestDataTime != nil {
		object["latestDataTime"] = test.NewObjectFromTime(*datum.LatestDataTime, objectFormat)
	}
	if datum.LastImportTime != nil {
		object["lastImportTime"] = test.NewObjectFromTime(*datum.LastImportTime, objectFormat)
	}
	return object
}

func MatchUpdate(datum *dataSource.Update) gomegaTypes.GomegaMatcher {
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"ProviderExternalID": gomega.Equal(datum.ProviderExternalID),
		"ProviderSessionID":  gomega.Equal(datum.ProviderSessionID),
		"State":              gomega.Equal(datum.State),
		"Metadata":           gomega.Equal(datum.Metadata),
		"Error":              gomega.Equal(datum.Error),
		"DataSetID":          gomega.Equal(datum.DataSetID),
		"EarliestDataTime":   test.MatchTime(datum.EarliestDataTime),
		"LatestDataTime":     test.MatchTime(datum.LatestDataTime),
		"LastImportTime":     test.MatchTime(datum.LastImportTime),
	}))
}

func RandomSource(options ...test.Option) *dataSource.Source {
	state := RandomState()
	datum := &dataSource.Source{}
	datum.ID = RandomDataSourceID()
	datum.UserID = userTest.RandomUserID()
	datum.ProviderType = authTest.RandomProviderType()
	datum.ProviderName = authTest.RandomProviderName()
	datum.ProviderExternalID = test.RandomOptional(authTest.RandomProviderExternalID, options...)
	datum.ProviderSessionID = test.Conditional(authTest.RandomProviderSessionID, state != dataSource.StateDisconnected)
	datum.State = state
	datum.Metadata = metadataTest.RandomMetadataMap()
	datum.Error = test.RandomOptionalPointer(errorsTest.RandomSerializable, options...)
	datum.DataSetID = test.RandomOptional(dataTest.RandomDataSetID, options...)
	datum.LastImportTime = test.RandomOptional(test.RandomTimeBeforeNow, options...)
	if datum.LastImportTime != nil && test.RandomBool() {
		datum.LatestDataTime = pointer.FromTime(test.RandomTimeBefore(*datum.LastImportTime))
		datum.EarliestDataTime = pointer.FromTime(test.RandomTimeBefore(*datum.LatestDataTime))
	}
	datum.CreatedTime = test.RandomTimeBefore(pointer.DefaultTime(datum.LastImportTime, time.Now()))
	datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(pointer.DefaultTime(datum.LastImportTime, datum.CreatedTime), time.Now()))
	datum.Revision = requestTest.RandomRevision()
	return datum
}

func CloneSource(datum *dataSource.Source) *dataSource.Source {
	if datum == nil {
		return nil
	}
	clone := &dataSource.Source{}
	clone.ID = datum.ID
	clone.UserID = datum.UserID
	clone.ProviderType = datum.ProviderType
	clone.ProviderName = datum.ProviderName
	clone.ProviderExternalID = pointer.CloneString(datum.ProviderExternalID)
	clone.ProviderSessionID = pointer.CloneString(datum.ProviderSessionID)
	clone.State = datum.State
	clone.Metadata = metadataTest.CloneMetadataMap(datum.Metadata)
	clone.Error = errorsTest.CloneSerializable(datum.Error)
	clone.DataSetID = pointer.CloneString(datum.DataSetID)
	clone.EarliestDataTime = pointer.CloneTime(datum.EarliestDataTime)
	clone.LatestDataTime = pointer.CloneTime(datum.LatestDataTime)
	clone.LastImportTime = pointer.CloneTime(datum.LastImportTime)
	clone.CreatedTime = datum.CreatedTime
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	clone.Revision = datum.Revision
	return clone
}

func NewObjectFromSource(datum *dataSource.Source, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	object["id"] = test.NewObjectFromString(datum.ID, objectFormat)
	object["userId"] = test.NewObjectFromString(datum.UserID, objectFormat)
	object["providerType"] = test.NewObjectFromString(datum.ProviderType, objectFormat)
	object["providerName"] = test.NewObjectFromString(datum.ProviderName, objectFormat)
	if datum.ProviderExternalID != nil {
		object["providerExternalId"] = test.NewObjectFromString(*datum.ProviderExternalID, objectFormat)
	}
	if datum.ProviderSessionID != nil {
		object["providerSessionId"] = test.NewObjectFromString(*datum.ProviderSessionID, objectFormat)
	}
	object["state"] = test.NewObjectFromString(datum.State, objectFormat)
	if len(datum.Metadata) > 0 {
		object["metadata"] = metadataTest.NewObjectFromMetadataMap(datum.Metadata, objectFormat)
	}
	if datum.Error != nil {
		object["error"] = errorsTest.NewObjectFromSerializable(datum.Error, objectFormat)
	}
	if datum.DataSetID != nil {
		object["dataSetId"] = test.NewObjectFromString(*datum.DataSetID, objectFormat)
	}
	if datum.EarliestDataTime != nil {
		object["earliestDataTime"] = test.NewObjectFromTime(*datum.EarliestDataTime, objectFormat)
	}
	if datum.LatestDataTime != nil {
		object["latestDataTime"] = test.NewObjectFromTime(*datum.LatestDataTime, objectFormat)
	}
	if datum.LastImportTime != nil {
		object["lastImportTime"] = test.NewObjectFromTime(*datum.LastImportTime, objectFormat)
	}
	object["createdTime"] = test.NewObjectFromTime(datum.CreatedTime, objectFormat)
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	object["revision"] = test.NewObjectFromInt(datum.Revision, objectFormat)
	return object
}

func MatchSource(datum *dataSource.Source) gomegaTypes.GomegaMatcher {
	if datum == nil {
		return gomega.BeNil()
	}
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"ID":                 gomega.Equal(datum.ID),
		"UserID":             gomega.Equal(datum.UserID),
		"ProviderType":       gomega.Equal(datum.ProviderType),
		"ProviderName":       gomega.Equal(datum.ProviderName),
		"ProviderExternalID": gomega.Equal(datum.ProviderExternalID),
		"ProviderSessionID":  gomega.Equal(datum.ProviderSessionID),
		"State":              gomega.Equal(datum.State),
		"Metadata":           gomega.Equal(datum.Metadata),
		"Error":              gomega.Equal(datum.Error),
		"DataSetID":          gomega.Equal(datum.DataSetID),
		"EarliestDataTime":   test.MatchTime(datum.EarliestDataTime),
		"LatestDataTime":     test.MatchTime(datum.LatestDataTime),
		"LastImportTime":     test.MatchTime(datum.LastImportTime),
		"CreatedTime":        gomega.Equal(datum.CreatedTime),
		"ModifiedTime":       test.MatchTime(datum.ModifiedTime),
		"Revision":           gomega.Equal(datum.Revision),
	}))
}

func RandomSourceArray(minimumLength int, maximumLength int, options ...test.Option) dataSource.SourceArray {
	datum := make(dataSource.SourceArray, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomSource(options...)
	}
	return datum
}

func CloneSourceArray(datum dataSource.SourceArray) dataSource.SourceArray {
	if len(datum) == 0 {
		return datum
	}
	clone := dataSource.SourceArray{}
	for _, source := range datum {
		clone = append(clone, CloneSource(source))
	}
	return clone
}

func MatchSourceArray(datum dataSource.SourceArray) gomegaTypes.GomegaMatcher {
	matchers := []gomegaTypes.GomegaMatcher{}
	for _, d := range datum {
		matchers = append(matchers, MatchSource(d))
	}
	return test.MatchArray(matchers)
}
