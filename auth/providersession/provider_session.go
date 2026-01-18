package providersession

import "github.com/tidepool-org/platform/auth"

//go:generate mockgen -destination=test/provider_session_mocks.go -package=test --mock_names=ProviderSessionAccessor=MockClient github.com/tidepool-org/platform/auth ProviderSessionAccessor
type Client auth.ProviderSessionAccessor
