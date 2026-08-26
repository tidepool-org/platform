package service_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	authServiceService "github.com/tidepool-org/platform/auth/service/service"
	authStoreTest "github.com/tidepool-org/platform/auth/store/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/provider"
	providerTest "github.com/tidepool-org/platform/provider/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Client", func() {
	var logger *logTest.Logger
	var authStore *authStoreTest.Store
	var providerFactory *providerTest.Factory
	var client *authServiceService.Client

	BeforeEach(func() {
		config := authClient.NewExternalConfig()
		config.Address = testHttp.NewAddress()
		config.UserAgent = testHttp.NewUserAgent()
		config.ServerSessionTokenSecret = authTest.NewServiceSecret()
		logger = logTest.NewLogger()
		authStore = authStoreTest.NewStore()
		providerFactory = providerTest.NewFactory()
		client = test.Must(authServiceService.NewClient(config, platform.AuthorizeAsService, test.RandomString(), logger, authStore, providerFactory))
	})

	AfterEach(func() {
		providerFactory.Expectations()
		authStore.Expectations()
	})

	Context("NewClient", func() {
		It("returns successfully", func() {
			Expect(client).ToNot(BeNil())
		})
	})

	Context("CreateProviderSession", func() {
		var ctx context.Context
		var create *auth.ProviderSessionCreate
		var providerSession *auth.ProviderSession
		var prvdr *slowCreateProvider
		var repository *authStoreTest.ProviderSessionRepository

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logger)
			create = authTest.RandomProviderSessionCreate()
			providerSession = authTest.RandomProviderSession()
			providerSession.Type = create.Type
			providerSession.Name = create.Name
			prvdr = newSlowCreateProvider(create.Type, create.Name)
			repository = authStore.NewProviderSessionRepositoryImpl
		})

		AfterEach(func() {
			prvdr.Expectations()
		})

		It("returns an error when the context is missing", func() {
			actualProviderSession, err := client.CreateProviderSession(context.Context(nil), create)
			errorsTest.ExpectEqual(err, errors.New("context is missing"))
			Expect(actualProviderSession).To(BeNil())
		})

		It("returns an error when the create is missing", func() {
			actualProviderSession, err := client.CreateProviderSession(ctx, nil)
			errorsTest.ExpectEqual(err, errors.New("create is missing"))
			Expect(actualProviderSession).To(BeNil())
		})

		It("returns an error when the provider factory has no provider for the create", func() {
			providerFactory.GetOutputs = []provider.Provider{nil}
			actualProviderSession, err := client.CreateProviderSession(ctx, create)
			errorsTest.ExpectEqual(err, errors.New("unable to get provider for provider session"))
			Expect(actualProviderSession).To(BeNil())
		})

		Context("with a provider for the create", func() {
			BeforeEach(func() {
				providerFactory.GetOutputs = []provider.Provider{prvdr}
				repository.CreateProviderSessionOutputs = []authTest.CreateProviderSessionOutput{{ProviderSession: providerSession}}
			})

			It("returns an error when the repository fails to create the provider session", func() {
				repositoryError := errorsTest.RandomError()
				repository.CreateProviderSessionOutputs = []authTest.CreateProviderSessionOutput{{Error: repositoryError}}
				actualProviderSession, err := client.CreateProviderSession(ctx, create)
				errorsTest.ExpectEqual(err, repositoryError)
				Expect(actualProviderSession).To(BeNil())
			})

			It("returns the provider session when the provider finalizes the creation", func() {
				prvdr.OnCreateOutputs = []error{nil}
				Expect(client.CreateProviderSession(ctx, create)).To(Equal(providerSession))
			})

			Context("when the provider fails to finalize the creation", func() {
				var onCreateError error

				BeforeEach(func() {
					onCreateError = errorsTest.RandomError()
					prvdr.OnCreateOutputs = []error{onCreateError}

					// The rollback resolves the provider a second time
					providerFactory.GetOutputs = append(providerFactory.GetOutputs, prvdr)
				})

				It("deletes the provider session and returns the creation error", func() {
					prvdr.OnDeleteOutputs = []error{nil}
					repository.DeleteProviderSessionOutputs = []error{nil}
					actualProviderSession, err := client.CreateProviderSession(ctx, create)
					errorsTest.ExpectEqual(err, onCreateError)
					Expect(actualProviderSession).To(BeNil())
					Expect(prvdr.OnDeleteInvocations).To(Equal(1))
					Expect(repository.DeleteProviderSessionInputs).To(HaveLen(1))
					Expect(repository.DeleteProviderSessionInputs[0].ID).To(Equal(providerSession.ID))
				})

				It("leaves the provider session when the provider fails to finalize the deletion", func() {
					prvdr.OnDeleteOutputs = []error{errorsTest.RandomError()}
					actualProviderSession, err := client.CreateProviderSession(ctx, create)
					errorsTest.ExpectEqual(err, onCreateError)
					Expect(actualProviderSession).To(BeNil())
					Expect(repository.DeleteProviderSessionInvocations).To(BeZero())
				})

				It("gives the deletion a timeout budget the failed creation did not consume", func() {
					prvdr.onCreateDuration = 100 * time.Millisecond
					prvdr.OnDeleteOutputs = []error{nil}
					repository.DeleteProviderSessionOutputs = []error{nil}
					_, err := client.CreateProviderSession(ctx, create)
					errorsTest.ExpectEqual(err, onCreateError)
					createDeadline := deadlineFromContext(prvdr.OnCreateInputs[0].Context)
					deleteDeadline := deadlineFromContext(prvdr.OnDeleteInputs[0].Context)
					Expect(deleteDeadline).To(BeTemporally(">=", createDeadline.Add(prvdr.onCreateDuration)))
				})
			})
		})
	})
})

// slowCreateProvider spends onCreateDuration inside OnCreate so that a spec can distinguish the deletion timeout
// budget from the creation timeout budget.
type slowCreateProvider struct {
	*providerTest.Provider
	onCreateDuration time.Duration
}

func newSlowCreateProvider(typ string, name string) *slowCreateProvider {
	return &slowCreateProvider{
		Provider: providerTest.NewProvider(typ, name),
	}
}

func (s *slowCreateProvider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	time.Sleep(s.onCreateDuration)
	return s.Provider.OnCreate(ctx, providerSession)
}

func deadlineFromContext(ctx context.Context) time.Time {
	deadline, ok := ctx.Deadline()
	Expect(ok).To(BeTrue())
	return deadline
}
