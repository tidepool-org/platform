package test

import (
	"github.com/tidepool-org/platform/association"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

func RandomAssociation() *association.Association {
	tipe := RandomType()
	datum := association.NewAssociation()
	switch tipe {
	case association.TypeDatum:
		datum.ID = pointer.FromString(dataTest.RandomID())
	}
	datum.Reason = pointer.FromString(RandomReason())
	datum.Type = pointer.FromString(tipe)
	switch tipe {
	case association.TypeURL:
		datum.URL = pointer.FromString(testHttp.NewURLString())
	}
	return datum
}

func CloneAssociation(datum *association.Association) *association.Association {
	if datum == nil {
		return nil
	}
	clone := association.NewAssociation()
	clone.ID = pointer.CloneString(datum.ID)
	clone.Reason = pointer.CloneString(datum.Reason)
	clone.Type = pointer.CloneString(datum.Type)
	clone.URL = pointer.CloneString(datum.URL)
	return clone
}

func NewObjectFromAssociation(datum *association.Association, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, objectFormat)
	}
	if datum.Reason != nil {
		object["reason"] = test.NewObjectFromString(*datum.Reason, objectFormat)
	}
	if datum.Type != nil {
		object["type"] = test.NewObjectFromString(*datum.Type, objectFormat)
	}
	if datum.URL != nil {
		object["url"] = test.NewObjectFromString(*datum.URL, objectFormat)
	}
	return object
}

func RandomType() string {
	return test.RandomStringFromArray(association.Types())
}

func RandomReason() string {
	return test.RandomStringFromRange(1, association.ReasonLengthMaximum)
}

func RandomAssociationArray() *association.AssociationArray {
	datumArray := association.NewAssociationArray()
	for count := test.RandomIntFromRange(1, 3); count > 0; count-- {
		*datumArray = append(*datumArray, RandomAssociation())
	}
	return datumArray
}

func CloneAssociationArray(datumArray *association.AssociationArray) *association.AssociationArray {
	if datumArray == nil {
		return nil
	}
	cloneArray := association.NewAssociationArray()
	for _, datum := range *datumArray {
		*cloneArray = append(*cloneArray, CloneAssociation(datum))
	}
	return cloneArray
}

func NewArrayFromAssociationArray(datumArray *association.AssociationArray, objectFormat test.ObjectFormat) []interface{} {
	if datumArray == nil {
		return nil
	}
	array := []interface{}{}
	for _, datum := range *datumArray {
		array = append(array, NewObjectFromAssociation(datum, objectFormat))
	}
	return array
}
