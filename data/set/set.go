package set

import "github.com/tidepool-org/platform/data"

//go:generate mockgen -destination=test/set_mocks.go -package=test --mock_names=DataSetAccessor=MockClient github.com/tidepool-org/platform/data DataSetAccessor
type Client data.DataSetAccessor
