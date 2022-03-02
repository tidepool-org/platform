package test

import (
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomIssue() *dataTypesDosingDecision.Issue {
	datum := dataTypesDosingDecision.NewIssue()
	datum.ID = pointer.FromString(test.RandomStringFromRange(1, dataTypesDosingDecision.IssueIDLengthMaximum))
	datum.Metadata = metadataTest.RandomMetadata()
	return datum
}

func CloneIssue(datum *dataTypesDosingDecision.Issue) *dataTypesDosingDecision.Issue {
	if datum == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewIssue()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Metadata = metadataTest.CloneMetadata(datum.Metadata)
	return clone
}

func NewObjectFromIssue(datum *dataTypesDosingDecision.Issue, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, objectFormat)
	}
	if datum.Metadata != nil {
		object["metadata"] = metadataTest.NewObjectFromMetadata(datum.Metadata, objectFormat)
	}
	return object
}

func RandomIssueArray() *dataTypesDosingDecision.IssueArray {
	datum := dataTypesDosingDecision.NewIssueArray()
	for count := 0; count < test.RandomIntFromRange(1, 3); count++ {
		*datum = append(*datum, RandomIssue())
	}
	return datum
}

func CloneIssueArray(datumArray *dataTypesDosingDecision.IssueArray) *dataTypesDosingDecision.IssueArray {
	if datumArray == nil {
		return nil
	}
	clone := dataTypesDosingDecision.NewIssueArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneIssue(datum))
	}
	return clone
}

func NewArrayFromIssueArray(datumArray *dataTypesDosingDecision.IssueArray, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromIssue(datum, objectFormat))
	}
	return array
}

func AnonymizeIssueArray(datumArray *dataTypesDosingDecision.IssueArray) []interface{} {
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, datum)
	}
	return array
}
