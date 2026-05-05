package test

import (
	"bytes"
	"strings"
	"time"

	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"

	compressTest "github.com/tidepool-org/platform/compress/test"
	crypto "github.com/tidepool-org/platform/crypto"
	dataRawStoreStructuredMongo "github.com/tidepool-org/platform/data/raw/store/structured/mongo"
	dataTest "github.com/tidepool-org/platform/data/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredTest "github.com/tidepool-org/platform/store/structured/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomDataRawID() string {
	return RandomDataRawIDFromTime(test.RandomTime())
}

func RandomDataRawIDFromTime(tm time.Time) string {
	return strings.Join([]string{bsonPrimitive.NewObjectID().Hex(), tm.Format(dataRawStoreStructuredMongo.IDDateFormat)}, dataRawStoreStructuredMongo.IDSeparator)
}

func RandomDocument(options ...test.Option) *dataRawStoreStructuredMongo.Document {
	return RandomDocumentWithCompressed(test.RandomBool(), options...)
}

func RandomDocumentWithUserIDAndDataSetID(userID string, dataSetID string, options ...test.Option) *dataRawStoreStructuredMongo.Document {
	document := RandomDocumentWithCompressed(test.RandomBool(), options...)
	document.UserID = userID
	document.DataSetID = dataSetID
	return document
}

func RandomDocumentWithCompressed(compressed bool, options ...test.Option) *dataRawStoreStructuredMongo.Document {
	originalData := test.RandomBytes()

	var data []byte
	if compressed {
		data = compressTest.Compress(originalData)
	} else {
		data = originalData
	}

	archived := test.IsOptionalPresent(options...)
	modifiedTime := test.RandomTimeBeforeNow()
	archivedTime := test.RandomTimeBefore(modifiedTime)
	archivableTime := test.RandomTimeBefore(archivedTime)
	processedTime := test.RandomTimeBefore(archivableTime)
	createdTime := test.RandomTimeBefore(processedTime)

	return &dataRawStoreStructuredMongo.Document{
		ID:             bsonPrimitive.NewObjectID(),
		UserID:         userTest.RandomUserID(),
		DataSetID:      dataTest.RandomDataSetID(),
		Metadata:       metadataTest.RandomMetadataMap(),
		DigestMD5:      crypto.Base64EncodedMD5Hash(originalData),
		DigestSHA256:   test.RandomOptional(func() string { return crypto.Base64EncodedSHA256Hash(originalData) }, options...),
		MediaType:      netTest.RandomMediaType(),
		Size:           len(originalData),
		Compressed:     compressed,
		Data:           bsonPrimitive.Binary{Data: data},
		ProcessedTime:  test.RandomOptional(test.Constant(processedTime), options...),
		ArchivableTime: test.Conditional(test.Constant(archivableTime), archived || test.RandomBool()),
		ArchivedTime:   test.Conditional(test.Constant(archivedTime), archived),
		CreatedTime:    createdTime,
		ModifiedTime:   test.RandomOptional(test.Constant(modifiedTime), options...),
		Revision:       storeStructuredTest.RandomRevision(),
	}
}

func CloneDocument(datum *dataRawStoreStructuredMongo.Document) *dataRawStoreStructuredMongo.Document {
	if datum == nil {
		return nil
	}
	return &dataRawStoreStructuredMongo.Document{
		ID:             datum.ID,
		UserID:         datum.UserID,
		DataSetID:      datum.DataSetID,
		Metadata:       metadataTest.CloneMetadataMap(datum.Metadata),
		DigestMD5:      datum.DigestMD5,
		DigestSHA256:   pointer.Clone(datum.DigestSHA256),
		MediaType:      datum.MediaType,
		Size:           datum.Size,
		Compressed:     datum.Compressed,
		Data:           bsonPrimitive.Binary{Data: bytes.Clone(datum.Data.Data)},
		ProcessedTime:  pointer.CloneTime(datum.ProcessedTime),
		ArchivableTime: pointer.CloneTime(datum.ArchivableTime),
		ArchivedTime:   pointer.CloneTime(datum.ArchivedTime),
		CreatedTime:    datum.CreatedTime,
		ModifiedTime:   pointer.CloneTime(datum.ModifiedTime),
		Revision:       datum.Revision,
	}
}
