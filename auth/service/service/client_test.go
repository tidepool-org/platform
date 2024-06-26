package service_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/auth/service/service"
	"github.com/tidepool-org/platform/auth/store"
	platformclient "github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/devicetokens"
	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
)

var _ = Describe("Client", func() {
	var testUserID = "test-user-id"
	var testDeviceToken1 = &devicetokens.DeviceToken{
		Apple: &devicetokens.AppleDeviceToken{
			Token:       []byte("test"),
			Environment: "sandbox",
		},
	}

	newTestServiceClient := func(url string, authStore store.Store) *service.Client {
		var err error
		extCfg := &client.ExternalConfig{
			Config: &platform.Config{
				Config: &platformclient.Config{
					Address:   url,
					UserAgent: "test",
				},
				ServiceSecret: "",
			},
			ServerSessionTokenSecret:  "test token",
			ServerSessionTokenTimeout: time.Minute,
		}
		authAs := platform.AuthorizeAsService
		name := "test auth client"
		logger := logtest.NewLogger()
		if authStore == nil {
			authStore = &mockAuthStore{
				DeviceTokenRepository: &mockDeviceTokenRepository{
					Tokens: map[string][]*devicetokens.DeviceToken{
						testUserID: {
							testDeviceToken1,
						},
					},
				},
			}
		}
		providerFactory := &mockProviderFactory{}
		serviceClient, err := service.NewClient(extCfg, authAs, name, logger, authStore, providerFactory)
		Expect(err).To(Succeed())
		return serviceClient
	}

	Describe("GetDeviceTokens", func() {
		It("returns a slice of tokens", func() {
			ctx := context.Background()
			server := NewServer()
			defer server.Close()
			serviceClient := newTestServiceClient(server.URL(), nil)

			tokens, err := serviceClient.GetDeviceTokens(ctx, testUserID)

			Expect(err).To(Succeed())
			Expect(tokens).To(HaveLen(1))
			Expect(tokens[0]).To(Equal(testDeviceToken1))
		})

		It("handles errors from the underlying repo", func() {
			ctx := context.Background()
			server := NewServer()
			defer server.Close()
			authStore := &mockAuthStore{
				DeviceTokenRepository: &mockDeviceTokenRepository{
					Error: fmt.Errorf("test error"),
				},
			}
			serviceClient := newTestServiceClient(server.URL(), authStore)

			_, err := serviceClient.GetDeviceTokens(ctx, testUserID)

			Expect(err).To(HaveOccurred())
		})
	})
})

type mockAuthStore struct {
	store.DeviceTokenRepository
}

func (s *mockAuthStore) NewProviderSessionRepository() store.ProviderSessionRepository {
	return nil
}

func (s *mockAuthStore) NewRestrictedTokenRepository() store.RestrictedTokenRepository {
	return nil
}

func (s *mockAuthStore) NewDeviceTokenRepository() store.DeviceTokenRepository {
	return s.DeviceTokenRepository
}

type mockProviderFactory struct{}

func (f *mockProviderFactory) Get(typ string, name string) (provider.Provider, error) {
	return nil, nil
}

type mockDeviceTokenRepository struct {
	Error  error
	Tokens map[string][]*devicetokens.DeviceToken
}

func (r *mockDeviceTokenRepository) GetAllByUserID(ctx context.Context, userID string) ([]*devicetokens.Document, error) {
	if r.Error != nil {
		return nil, r.Error
	}

	if tokens, ok := r.Tokens[userID]; ok {
		docs := make([]*devicetokens.Document, 0, len(tokens))
		for _, token := range tokens {
			docs = append(docs, &devicetokens.Document{DeviceToken: *token})
		}
		return docs, nil
	}
	return nil, nil
}

func (r *mockDeviceTokenRepository) Upsert(ctx context.Context, doc *devicetokens.Document) error {
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func (r *mockDeviceTokenRepository) EnsureIndexes() error {
	if r.Error != nil {
		return r.Error
	}
	return nil
}
