package set

import "github.com/tidepool-org/platform/data"

//go:generate go tool go.uber.org/mock/mockgen -destination=test/set_mocks.go -package=test -typed -mock_names=DataSetAccessor=MockClient github.com/tidepool-org/platform/data DataSetAccessor

type Client data.DataSetAccessor
