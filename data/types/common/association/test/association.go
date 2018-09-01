package test

import (
	"math/rand"

	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/common/association"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

func NewAssociation() *association.Association {
	typ := test.RandomStringFromArray(association.Types())
	datum := association.NewAssociation()
	if typ == association.TypeDatum {
		datum.ID = pointer.FromString(dataTest.RandomID())
	}
	datum.Reason = pointer.FromString(test.NewText(1, 1000))
	datum.Type = pointer.FromString(typ)
	if typ == association.TypeURL {
		datum.URL = pointer.FromString(testHTTP.NewURLString())
	}
	return datum
}

func CloneAssociation(datum *association.Association) *association.Association {
	if datum == nil {
		return nil
	}
	clone := association.NewAssociation()
	clone.ID = test.CloneString(datum.ID)
	clone.Reason = test.CloneString(datum.Reason)
	clone.Type = test.CloneString(datum.Type)
	clone.URL = test.CloneString(datum.URL)
	return clone
}

func NewAssociationArray() *association.AssociationArray {
	datum := association.NewAssociationArray()
	for count := rand.Intn(3); count >= 0; count-- {
		*datum = append(*datum, NewAssociation())
	}
	return datum
}

func CloneAssociationArray(datumArray *association.AssociationArray) *association.AssociationArray {
	if datumArray == nil {
		return nil
	}
	clone := association.NewAssociationArray()
	for _, datum := range *datumArray {
		*clone = append(*clone, CloneAssociation(datum))
	}
	return clone
}
