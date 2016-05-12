package types

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/sample"
	"github.com/tidepool-org/platform/pvn/data/types/base/sample/sub"
	"github.com/tidepool-org/platform/pvn/data/types/base/upload"
)

func Parse(context data.Context, parser data.ObjectParser) (data.Datum, error) {
	if context == nil {
		return nil, app.Error("types", "context is missing")
	}
	if parser == nil {
		return nil, app.Error("types", "parser is missing")
	}

	var datum data.Datum

	datumType := parser.ParseString("type")
	if datumType == nil {
		context.AppendError("type", ErrorValueMissing())
		return nil, nil
	}

	datumSubType := parser.ParseString("subType")

	// NOTE: This is the "master switchboard" that creates all of the datum from
	// the type and subType

	switch *datumType {
	case sample.Type():
		if datumSubType != nil {
			switch *datumSubType {
			case sub.SubType():
				datum = sub.New()
			default:
				context.AppendError("subType", ErrorSubTypeInvalid(*datumSubType))
				return nil, nil
			}
		} else {
			datum = sample.New()
		}
	case upload.Type():
		return upload.New()
	default:
		context.AppendError("type", ErrorTypeInvalid(*datumType))
		return nil, nil
	}

	datum.Parse(parser)

	return datum, nil
}
