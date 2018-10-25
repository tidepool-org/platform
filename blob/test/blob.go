package test

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	requestTest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomID() string {
	return blob.NewID()
}

func RandomStatuses() []string {
	return test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(blob.Statuses()), blob.Statuses())
}

func RandomFilter() *blob.Filter {
	datum := &blob.Filter{}
	datum.MediaType = pointer.FromStringArray(netTest.RandomMediaTypes(1, 3))
	datum.Status = pointer.FromStringArray(RandomStatuses())
	return datum
}

func NewObjectFromFilter(datum *blob.Filter, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.MediaType != nil {
		object["mediaType"] = test.NewObjectFromStringArray(*datum.MediaType, objectFormat)
	}
	if datum.Status != nil {
		object["status"] = test.NewObjectFromStringArray(*datum.Status, objectFormat)
	}
	return object
}

func RandomContent() *blob.Content {
	content := test.RandomBytes()
	datum := &blob.Content{}
	datum.Body = ioutil.NopCloser(bytes.NewReader(content))
	datum.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(content))
	datum.MediaType = pointer.FromString(netTest.RandomMediaType())
	return datum
}

func RandomBlob() *blob.Blob {
	datum := &blob.Blob{}
	datum.ID = pointer.FromString(RandomID())
	datum.UserID = pointer.FromString(userTest.RandomID())
	datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
	datum.MediaType = pointer.FromString(netTest.RandomMediaType())
	datum.Size = pointer.FromInt(test.RandomIntFromRange(1, 100*1024*1024))
	datum.Status = pointer.FromString(test.RandomStringFromArray(blob.Statuses()))
	datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
	if *datum.Status == blob.StatusAvailable {
		datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
	}
	datum.Revision = pointer.FromInt(requestTest.RandomRevision())
	return datum
}

func CloneBlob(datum *blob.Blob) *blob.Blob {
	if datum == nil {
		return nil
	}
	clone := &blob.Blob{}
	clone.ID = pointer.CloneString(datum.ID)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.DigestMD5 = pointer.CloneString(datum.DigestMD5)
	clone.MediaType = pointer.CloneString(datum.MediaType)
	clone.Size = pointer.CloneInt(datum.Size)
	clone.Status = pointer.CloneString(datum.Status)
	clone.CreatedTime = pointer.CloneTime(datum.CreatedTime)
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	clone.Revision = test.CloneInt(datum.Revision)
	return clone
}

func NewObjectFromBlob(datum *blob.Blob, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, objectFormat)
	}
	if datum.UserID != nil {
		object["userId"] = test.NewObjectFromString(*datum.UserID, objectFormat)
	}
	if datum.DigestMD5 != nil {
		object["digestMD5"] = test.NewObjectFromString(*datum.DigestMD5, objectFormat)
	}
	if datum.MediaType != nil {
		object["mediaType"] = test.NewObjectFromString(*datum.MediaType, objectFormat)
	}
	if datum.Size != nil {
		object["size"] = test.NewObjectFromInt(*datum.Size, objectFormat)
	}
	if datum.Status != nil {
		object["status"] = test.NewObjectFromString(*datum.Status, objectFormat)
	}
	if datum.CreatedTime != nil {
		object["createdTime"] = test.NewObjectFromTime(*datum.CreatedTime, objectFormat)
	}
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	if datum.Revision != nil {
		object["revision"] = test.NewObjectFromInt(*datum.Revision, objectFormat)
	}
	return object
}

func ExpectEqualBlob(actual *blob.Blob, expected *blob.Blob) {
	gomega.Expect(actual).ToNot(gomega.BeNil())
	gomega.Expect(expected).ToNot(gomega.BeNil())
	gomega.Expect(actual.ID).To(gomega.Equal(expected.ID))
	gomega.Expect(actual.UserID).To(gomega.Equal(expected.UserID))
	gomega.Expect(actual.DigestMD5).To(gomega.Equal(expected.DigestMD5))
	gomega.Expect(actual.MediaType).To(gomega.Equal(expected.MediaType))
	gomega.Expect(actual.Size).To(gomega.Equal(expected.Size))
	gomega.Expect(actual.Status).To(gomega.Equal(expected.Status))
	if actual.CreatedTime != nil && expected.CreatedTime != nil {
		gomega.Expect(actual.CreatedTime.Local()).To(gomega.Equal(expected.CreatedTime.Local()))
	} else {
		gomega.Expect(actual.CreatedTime).To(gomega.Equal(expected.CreatedTime))
	}
	if actual.ModifiedTime != nil && expected.ModifiedTime != nil {
		gomega.Expect(actual.ModifiedTime.Local()).To(gomega.Equal(expected.ModifiedTime.Local()))
	} else {
		gomega.Expect(actual.ModifiedTime).To(gomega.Equal(expected.ModifiedTime))
	}
	gomega.Expect(actual.Revision).To(gomega.Equal(expected.Revision))
}

func RandomBlobs(minimumLength int, maximumLength int) blob.Blobs {
	datum := make(blob.Blobs, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomBlob()
	}
	return datum
}

func ExpectEqualBlobs(actual blob.Blobs, expected blob.Blobs) {
	gomega.Expect(actual).To(gomega.HaveLen(len(expected)))
	for index := range expected {
		ExpectEqualBlob(actual[index], expected[index])
	}
}
