package data

import (
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

// TODO: Temporary until all data types properly built

type Any interface{}

func BuildAny(datum types.Datum, errs validate.ErrorProcessing) Any {
	return datum
}
