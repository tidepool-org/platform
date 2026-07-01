package providersession

import "github.com/tidepool-org/platform/auth"

//go:generate go tool go.uber.org/mock/mockgen -destination=test/provider_session_mocks.go -package=test -typed -mock_names=ProviderSessionClient=MockClient github.com/tidepool-org/platform/auth ProviderSessionClient

type Client auth.ProviderSessionClient
