package test

import (
	"math/rand"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

const (
	BlobValuesMaximum      = 3
	BlobArrayValuesMaximum = 3
)

func NewBlob() *data.Blob {
	datum := data.NewBlob()
	for index := rand.Intn(BlobValuesMaximum); index >= 0; index-- {
		(*datum)[test.NewVariableString(1, 8, test.CharsetAlpha)] = test.NewVariableString(0, 16, test.CharsetAlpha)
	}
	return datum
}

func CloneBlob(datum *data.Blob) *data.Blob {
	if datum == nil {
		return nil
	}
	clone := data.NewBlob()
	for key, value := range *datum {
		(*clone)[key] = value
	}
	return clone
}

func NewBlobArray() *data.BlobArray {
	datum := data.NewBlobArray()
	for index := rand.Intn(BlobArrayValuesMaximum); index >= 0; index-- {
		*datum = append(*datum, NewBlob())
	}
	return datum
}

func CloneBlobArray(datum *data.BlobArray) *data.BlobArray {
	if datum == nil {
		return nil
	}
	clone := data.NewBlobArray()
	for _, blob := range *datum {
		*clone = append(*clone, CloneBlob(blob))
	}
	return clone
}
