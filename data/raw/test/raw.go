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

func RandomFilter(options ...test.Option) *dataRaw.Filter {
	return &dataRaw.Filter{
		CreatedDate: test.RandomOptional(RandomCreatedDate, options...),
		DataSetID:   test.RandomOptional(dataTest.RandomDataSetID, options...),
		Processed:   test.RandomOptional(test.RandomBool, options...),
		Archivable:  test.RandomOptional(test.RandomBool, options...),
		Archived:    test.RandomOptional(test.RandomBool, options...),
	}
}

func CloneFilter(datum *dataRaw.Filter) *dataRaw.Filter {
	if datum == nil {
		return nil
	}
	return &dataRaw.Filter{
		CreatedDate: pointer.CloneString(datum.CreatedDate),
		DataSetID:   pointer.CloneString(datum.DataSetID),
		Processed:   pointer.CloneBool(datum.Processed),
		Archivable:  pointer.CloneBool(datum.Archivable),
		Archived:    pointer.CloneBool(datum.Archived),
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
	if datum.DataSetID != nil {
		object["dataSetId"] = test.NewObjectFromString(*datum.DataSetID, objectFormat)
	}
	if datum.Processed != nil {
		object["processed"] = test.NewObjectFromBool(*datum.Processed, objectFormat)
	}
	if datum.Archivable != nil {
		object["archivable"] = test.NewObjectFromBool(*datum.Archivable, objectFormat)
	}
	if datum.Archived != nil {
		object["archived"] = test.NewObjectFromBool(*datum.Archived, objectFormat)
	}
	return object
}

func RandomCreate(options ...test.Option) *dataRaw.Create {
	return &dataRaw.Create{
		Metadata:       metadataTest.RandomMetadataMap(),
		DigestMD5:      test.RandomOptional(netTest.RandomDigestMD5, options...),
		MediaType:      test.RandomOptional(netTest.RandomMediaType, options...),
		ArchivableTime: test.RandomOptional(test.RandomTimeBeforeNow, options...),
	}
}

func CloneCreate(datum *dataRaw.Create) *dataRaw.Create {
	if datum == nil {
		return nil
	}
	return &dataRaw.Create{
		Metadata:       metadataTest.CloneMetadataMap(datum.Metadata),
		DigestMD5:      pointer.CloneString(datum.DigestMD5),
		MediaType:      pointer.CloneString(datum.MediaType),
		ArchivableTime: pointer.CloneTime(datum.ArchivableTime),
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
	if datum.ArchivableTime != nil {
		object["archivableTime"] = test.NewObjectFromTime(*datum.ArchivableTime, objectFormat)
	}
	return object
}

func RandomContent(options ...test.Option) *dataRaw.Content {
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

func RandomUpdate(options ...test.Option) *dataRaw.Update {
	archived := !test.Options(options).AllowOptional() || test.RandomBool()
	archivedTime := test.RandomTimeBeforeNow()
	archivableTime := test.RandomTimeBefore(archivedTime)
	processedTime := test.RandomTimeBefore(archivableTime)

	for {
		update := &dataRaw.Update{
			ProcessedTime:  test.RandomOptional(test.Constant(processedTime), options...),
			ArchivableTime: test.Conditional(test.Constant(archivableTime), archived || test.RandomBool()),
			ArchivedTime:   test.Conditional(test.Constant(archivedTime), archived),
			Metadata:       test.RandomOptionalPointer(metadataTest.RandomMetadataMapPointer, options...),
		}
		if update.ProcessedTime != nil || update.ArchivableTime != nil || update.ArchivedTime != nil || update.Metadata != nil {
			return update
		}
	}
}

func CloneUpdate(datum *dataRaw.Update) *dataRaw.Update {
	if datum == nil {
		return nil
	}
	return &dataRaw.Update{
		ProcessedTime:  pointer.CloneTime(datum.ProcessedTime),
		ArchivableTime: pointer.CloneTime(datum.ArchivableTime),
		ArchivedTime:   pointer.CloneTime(datum.ArchivedTime),
		Metadata:       metadataTest.CloneMetadataMapPointer(datum.Metadata),
	}
}

func NewObjectFromUpdate(datum *dataRaw.Update, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.ProcessedTime != nil {
		object["processedTime"] = test.NewObjectFromTime(*datum.ProcessedTime, objectFormat)
	}
	if datum.ArchivableTime != nil {
		object["archivableTime"] = test.NewObjectFromTime(*datum.ArchivableTime, objectFormat)
	}
	if datum.ArchivedTime != nil {
		object["archivedTime"] = test.NewObjectFromTime(*datum.ArchivedTime, objectFormat)
	}
	if datum.Metadata != nil {
		object["metadata"] = metadataTest.NewObjectFromMetadataMap(*datum.Metadata, objectFormat)
	}
	return object
}

func RandomRaw(options ...test.Option) *dataRaw.Raw {
	archived := !test.Options(options).AllowOptional() || test.RandomBool()
	modifiedTime := test.RandomTimeBeforeNow()
	archivedTime := test.RandomTimeBefore(modifiedTime)
	archivableTime := test.RandomTimeBefore(archivedTime)
	processedTime := test.RandomTimeBefore(archivableTime)
	createdTime := test.RandomTimeBefore(processedTime)

	return &dataRaw.Raw{
		ID:             RandomDataRawIDFromTime(createdTime),
		UserID:         userTest.RandomUserID(),
		DataSetID:      dataTest.RandomDataSetID(),
		Metadata:       metadataTest.RandomMetadataMap(),
		DigestMD5:      netTest.RandomDigestMD5(),
		MediaType:      netTest.RandomMediaType(),
		Size:           test.RandomIntFromRange(0, 1024),
		ProcessedTime:  test.RandomOptional(test.Constant(processedTime), options...),
		ArchivableTime: test.Conditional(test.Constant(processedTime), archived || test.RandomBool()),
		ArchivedTime:   test.Conditional(test.Constant(processedTime), archived),
		CreatedTime:    createdTime,
		ModifiedTime:   test.RandomOptional(test.Constant(modifiedTime), options...),
		Revision:       storeStructuredTest.RandomRevision(),
	}
}

func CloneRaw(datum *dataRaw.Raw) *dataRaw.Raw {
	if datum == nil {
		return nil
	}
	return &dataRaw.Raw{
		ID:             datum.ID,
		UserID:         datum.UserID,
		DataSetID:      datum.DataSetID,
		Metadata:       metadataTest.CloneMetadataMap(datum.Metadata),
		DigestMD5:      datum.DigestMD5,
		MediaType:      datum.MediaType,
		Size:           datum.Size,
		ProcessedTime:  pointer.CloneTime(datum.ProcessedTime),
		ArchivableTime: pointer.CloneTime(datum.ArchivableTime),
		ArchivedTime:   pointer.CloneTime(datum.ArchivedTime),
		CreatedTime:    datum.CreatedTime,
		ModifiedTime:   pointer.CloneTime(datum.ModifiedTime),
		Revision:       datum.Revision,
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
	if datum.Metadata != nil {
		object["metadata"] = metadataTest.NewObjectFromMetadataMap(datum.Metadata, objectFormat)
	}
	object["digestMD5"] = test.NewObjectFromString(datum.DigestMD5, objectFormat)
	object["mediaType"] = test.NewObjectFromString(datum.MediaType, objectFormat)
	object["size"] = test.NewObjectFromInt(datum.Size, objectFormat)
	if datum.ProcessedTime != nil {
		object["processedTime"] = test.NewObjectFromTime(*datum.ProcessedTime, objectFormat)
	}
	if datum.ArchivableTime != nil {
		object["archivableTime"] = test.NewObjectFromTime(*datum.ArchivableTime, objectFormat)
	}
	if datum.ArchivedTime != nil {
		object["archivedTime"] = test.NewObjectFromTime(*datum.ArchivedTime, objectFormat)
	}
	object["createdTime"] = test.NewObjectFromTime(datum.CreatedTime, objectFormat)
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	object["revision"] = test.NewObjectFromInt(datum.Revision, objectFormat)
	return object
}
