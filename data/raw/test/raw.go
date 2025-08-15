package test

import (
	"time"

	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawStoreStructuredMongoTest "github.com/tidepool-org/platform/data/raw/store/structured/mongo/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredTest "github.com/tidepool-org/platform/store/structured/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomDataRawID() string {
	return dataRawStoreStructuredMongoTest.RandomDataRawID()
}

func RandomDataRawIDFromTime(tm time.Time) string {
	return dataRawStoreStructuredMongoTest.RandomDataRawIDFromTime(tm)
}

func RandomCreatedDate() string {
	return test.RandomTimeBeforeNow().UTC().Format(dataRaw.FilterCreatedDateFormat)
}

func RandomFilter() *dataRaw.Filter {
	return &dataRaw.Filter{
		CreatedDate: pointer.FromString(RandomCreatedDate()),
		DataSetIDs:  pointer.FromStringArray(test.RandomStringArrayFromRangeAndGeneratorWithDuplicates(1, 3, dataTest.RandomDataSetID)),
		Processed:   pointer.FromBool(test.RandomBool()),
	}
}

func CloneFilter(datum *dataRaw.Filter) *dataRaw.Filter {
	if datum == nil {
		return nil
	}
	return &dataRaw.Filter{
		CreatedDate: pointer.CloneString(datum.CreatedDate),
		DataSetIDs:  pointer.CloneStringArray(datum.DataSetIDs),
		Processed:   pointer.CloneBool(datum.Processed),
	}
}

func NewObjectFromFilter(datum *dataRaw.Filter, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.CreatedDate != nil {
		object["createdDate"] = test.NewObjectFromString(*datum.CreatedDate, objectFormat)
	}
	if datum.DataSetIDs != nil {
		object["dataSetIds"] = test.NewObjectFromStringArray(*datum.DataSetIDs, objectFormat)
	}
	if datum.Processed != nil {
		object["processed"] = test.NewObjectFromBool(*datum.Processed, objectFormat)
	}
	return object
}

func RandomCreate() *dataRaw.Create {
	return &dataRaw.Create{
		Metadata:  metadataTest.RandomMetadataMap(),
		DigestMD5: pointer.FromString(netTest.RandomDigestMD5()),
		MediaType: pointer.FromString(netTest.RandomMediaType()),
	}
}

func CloneCreate(datum *dataRaw.Create) *dataRaw.Create {
	if datum == nil {
		return nil
	}
	return &dataRaw.Create{
		Metadata:  metadataTest.CloneMetadataMap(datum.Metadata),
		DigestMD5: pointer.CloneString(datum.DigestMD5),
		MediaType: pointer.CloneString(datum.MediaType),
	}
}

func NewObjectFromCreate(datum *dataRaw.Create, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if len(datum.Metadata) > 0 {
		object["metadata"] = metadataTest.NewObjectFromMetadataMap(datum.Metadata, objectFormat)
	}
	if datum.DigestMD5 != nil {
		object["digestMD5"] = test.NewObjectFromString(*datum.DigestMD5, objectFormat)
	}
	if datum.MediaType != nil {
		object["mediaType"] = test.NewObjectFromString(*datum.MediaType, objectFormat)
	}
	return object
}

func RandomContent() *dataRaw.Content {
	return &dataRaw.Content{
		DigestMD5:  netTest.RandomDigestMD5(),
		MediaType:  netTest.RandomMediaType(),
		ReadCloser: test.RandomReadCloser(),
	}
}

func CloneContent(datum *dataRaw.Content) *dataRaw.Content {
	if datum == nil {
		return nil
	}
	return &dataRaw.Content{
		DigestMD5:  datum.DigestMD5,
		MediaType:  datum.MediaType,
		ReadCloser: datum.ReadCloser,
	}
}

func NewObjectFromContent(datum *dataRaw.Content, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	object["digestMD5"] = test.NewObjectFromString(datum.DigestMD5, objectFormat)
	object["mediaType"] = test.NewObjectFromString(datum.MediaType, objectFormat)
	return object
}

func RandomUpdate() *dataRaw.Update {
	return &dataRaw.Update{
		ProcessedTime: test.RandomTimeBeforeNow(),
	}
}

func CloneUpdate(datum *dataRaw.Update) *dataRaw.Update {
	if datum == nil {
		return nil
	}
	return &dataRaw.Update{
		ProcessedTime: datum.ProcessedTime,
	}
}

func NewObjectFromUpdate(datum *dataRaw.Update, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	object["processedTime"] = test.NewObjectFromTime(datum.ProcessedTime, objectFormat)
	return object
}

func RandomRaw() *dataRaw.Raw {
	createdTime := test.RandomTimeBeforeNow()
	return &dataRaw.Raw{
		ID:            RandomDataRawIDFromTime(createdTime),
		UserID:        userTest.RandomUserID(),
		DataSetID:     dataTest.RandomDataSetID(),
		Metadata:      metadataTest.RandomMetadataMap(),
		DigestMD5:     netTest.RandomDigestMD5(),
		MediaType:     netTest.RandomMediaType(),
		Size:          test.RandomIntFromRange(0, 1024),
		ProcessedTime: pointer.FromTime(test.RandomTimeAfter(createdTime)),
		CreatedTime:   createdTime,
		ModifiedTime:  pointer.FromTime(test.RandomTimeAfter(createdTime)),
		Revision:      storeStructuredTest.RandomRevision(),
	}
}

func CloneRaw(datum *dataRaw.Raw) *dataRaw.Raw {
	if datum == nil {
		return nil
	}
	return &dataRaw.Raw{
		ID:            datum.ID,
		UserID:        datum.UserID,
		DataSetID:     datum.DataSetID,
		Metadata:      metadataTest.CloneMetadataMap(datum.Metadata),
		DigestMD5:     datum.DigestMD5,
		MediaType:     datum.MediaType,
		Size:          datum.Size,
		ProcessedTime: pointer.CloneTime(datum.ProcessedTime),
		CreatedTime:   datum.CreatedTime,
		ModifiedTime:  pointer.CloneTime(datum.ModifiedTime),
		Revision:      datum.Revision,
	}
}

func NewObjectFromRaw(datum *dataRaw.Raw, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	object["id"] = test.NewObjectFromString(datum.ID, objectFormat)
	object["userId"] = test.NewObjectFromString(datum.UserID, objectFormat)
	object["dataSetId"] = test.NewObjectFromString(datum.DataSetID, objectFormat)
	if len(datum.Metadata) > 0 {
		object["metadata"] = metadataTest.NewObjectFromMetadataMap(datum.Metadata, objectFormat)
	}
	object["digestMD5"] = test.NewObjectFromString(datum.DigestMD5, objectFormat)
	object["mediaType"] = test.NewObjectFromString(datum.MediaType, objectFormat)
	object["size"] = test.NewObjectFromInt(datum.Size, objectFormat)
	if datum.ProcessedTime != nil {
		object["processedTime"] = test.NewObjectFromTime(*datum.ProcessedTime, objectFormat)
	}
	object["createdTime"] = test.NewObjectFromTime(datum.CreatedTime, objectFormat)
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	object["revision"] = test.NewObjectFromInt(datum.Revision, objectFormat)
	return object
}
