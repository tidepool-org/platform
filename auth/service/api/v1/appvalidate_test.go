package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	gomock "github.com/golang/mock/gomock"

	"github.com/tidepool-org/platform/appvalidate"
	"github.com/tidepool-org/platform/auth"
	v1 "github.com/tidepool-org/platform/auth/service/api/v1"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/middleware"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -build_flags=--mod=mod -destination=./auth_service_mock.go -package=v1 -mock_names Service=MockAuthService github.com/tidepool-org/platform/auth/service Service

var _ = Describe("App Validation", func() {
	defer GinkgoRecover()

	var ctrl *gomock.Controller
	var service *v1.MockAuthService
	var repo *appvalidate.MockRepository
	var generator *appvalidate.MockChallengeGenerator
	var authClient *authTest.MockClient
	var handler http.Handler

	challenge := "challenge"
	serverSessionToken := "serverToken"

	unattestedUser := user{
		UserID:              "unattested",
		SessionToken:        "unattestedToken",
		Details:             request.NewAuthDetails(request.MethodSessionToken, "unattested", "unattestedToken"),
		AttestationVerified: false,
	}
	attestedUser := user{
		UserID:               "attested",
		SessionToken:         "attestedToken",
		Details:              request.NewAuthDetails(request.MethodSessionToken, "attested", "attestedToken"),
		KeyID:                "YWJjZGVmYWJjZGVm",
		AttestationVerified:  false,
		AttestationChallenge: challenge,
	}
	attestedUnverifiedUser := user{
		UserID:               "attestedUnverified",
		SessionToken:         "attestedUnverifiedToken",
		Details:              request.NewAuthDetails(request.MethodSessionToken, "attestedUnverified", "attestedUnverified"),
		KeyID:                "YWRzZmFkZg==",
		AttestationVerified:  false,
		AttestationChallenge: challenge,
	}
	attestedVerifiedUser := user{
		UserID:               "attestedVerified",
		SessionToken:         "attestedVerifiedToken",
		Details:              request.NewAuthDetails(request.MethodSessionToken, "attestedVerified", "attestedVerifiedToken"),
		KeyID:                "YWJkZmRlZg=",
		AttestationVerified:  true,
		AttestationChallenge: challenge,
		AssertionChallenge:   challenge,
	}
	users := []user{
		unattestedUser,
		attestedUser,
		attestedVerifiedUser,
		attestedUnverifiedUser,
	}

	initialValidations := make([]appvalidate.AppValidation, len(users))
	for i, user := range users {
		validation := appvalidate.AppValidation{
			UserID:               user.UserID,
			KeyID:                user.KeyID,
			Verified:             user.AttestationVerified,
			AttestationChallenge: user.AttestationChallenge,
			AssertionChallenge:   user.AssertionChallenge,
		}
		if user.AttestationVerified {
			validation.AttestationVerifiedTime = pointer.FromTime(time.Date(2023, time.January, 3, 10, 0, 0, 0, time.UTC))
		}
		initialValidations[i] = validation
	}
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		service = v1.NewMockAuthService(ctrl)
		repo = newRepository(ctrl, initialValidations)
		generator = appvalidate.NewMockChallengeGenerator(ctrl)
		authClient = authTest.NewMockClient(ctrl)

		service.EXPECT().
			Logger().
			Return(logTest.NewLogger()).
			AnyTimes()
		generator.EXPECT().
			GenerateChallenge(gomock.Any()).
			Return(challenge, nil).
			AnyTimes()
		validator, err := appvalidate.NewValidator(repo, generator, appvalidate.ValidatorConfig{
			AppleAppIDs:   []string{"org.tidepool.app"},
			ChallengeSize: 10,
		})
		Expect(err).ToNot(HaveOccurred())

		service.EXPECT().
			AppValidator().
			Return(validator).
			AnyTimes()

		authClient.EXPECT().
			ServerSessionToken().
			Return(serverSessionToken, nil).
			AnyTimes()
		authClient.EXPECT().
			ValidateSessionToken(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, token string) (request.AuthDetails, error) {
				for _, user := range users {
					if token == user.SessionToken {
						return user.Details, nil
					}
				}
				return nil, request.ErrorUnauthorized()
			}).
			AnyTimes()

		api := rest.NewApi()

		router, err := v1.NewRouter(service)
		Expect(err).ToNot(HaveOccurred())

		authMiddleware, err := middleware.NewAuthenticator("secret", authClient)
		Expect(err).ToNot(HaveOccurred())

		// Use a subset of the middlewares used in the actual
		// API.InitializeMiddleware - just auth is really needed for testing.
		middlewares := []rest.Middleware{
			authMiddleware,
		}
		api.Use(middlewares...)

		app, err := rest.MakeRouter(router.Routes()...)
		if err != nil {
			Expect(err).ToNot(HaveOccurred())
		}
		api.SetApp(app)
		handler = api.MakeHandler()
	})

	Describe("POST /v1/attestations/challenges", func() {
		It("succeeds with correct input", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: "YWJjZGVmZ2hpamFiY2RlZmdoaWphYmNkZWZnaGlq",
			}
			req := newRequest(http.MethodPost, "/v1/attestations/challenges", unattestedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusCreated))
			resp := w.Result()
			var result appvalidate.ChallengeResult
			err := unmashalBody(resp.Body, &result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Challenge).To(Equal(challenge))
		})

		It("fails with empty keyID", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: "",
			}

			req := newRequest(http.MethodPost, "/v1/attestations/challenges", unattestedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("fails if unauthorized", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: "YWJjZGVmZ2hpamFiY2RlZmdoaWphYmNkZWZnaGlq",
			}

			noSessionToken := ""
			req := newRequest(http.MethodPost, "/v1/attestations/challenges", noSessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})

		It("fails with bad session token", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: "YWJjZGVmZ2hpamFiY2RlZmdoaWphYmNkZWZnaGlq",
			}

			badSessionToken := "BAD_TOKEN!"
			req := newRequest(http.MethodPost, "/v1/attestations/challenges", badSessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("POST /v1/assertions/challenges", func() {
		It("fails with an unverified user", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: attestedUnverifiedUser.KeyID,
			}

			req := newRequest(http.MethodPost, "/v1/assertions/challenges", attestedUnverifiedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("succeeds only with a verified attested user", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: attestedVerifiedUser.KeyID,
			}

			req := newRequest(http.MethodPost, "/v1/assertions/challenges", attestedVerifiedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusCreated))
			resp := w.Result()
			var result appvalidate.ChallengeResult
			err := unmashalBody(resp.Body, &result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Challenge).To(Equal(challenge))
		})

		It("fails with empty keyID", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: "",
			}

			req := newRequest(http.MethodPost, "/v1/assertions/challenges", unattestedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("fails if unauthorized", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: "YWJjZGVmZ2hpamFiY2RlZmdoaWphYmNkZWZnaGlq",
			}

			noSessionToken := ""
			req := newRequest(http.MethodPost, "/v1/assertions/challenges", noSessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})

		It("fails with bad session token", func() {
			body := &appvalidate.ChallengeCreate{
				KeyID: "YWJjZGVmZ2hpamFiY2RlZmdoaWphYmNkZWZnaGlq",
			}

			badSessionToken := "BAD_TOKEN!"
			req := newRequest(http.MethodPost, "/v1/assertions/challenges", badSessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("POST /v1/attestations/verifications", func() {
		// Was going to use an actual signed object from apple
		// but unfortunately the expiration time for that is only
		// a few days so there is no integration test for that.
		It("fails on attestation that is not base64 encoded", func() {
			body := &appvalidate.AttestationVerify{
				KeyID:       attestedUser.KeyID,
				Challenge:   challenge,
				Attestation: `{"key": "field"}`,
			}

			req := newRequest(http.MethodPost, "/v1/attestations/verifications", attestedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
		It("fails on incorrect attestation", func() {
			body := &appvalidate.AttestationVerify{
				KeyID:       attestedUser.KeyID,
				Challenge:   challenge,
				Attestation: `YWJjZGVm`,
			}

			req := newRequest(http.MethodPost, "/v1/attestations/verifications", attestedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("POST /v1/assertions/verifications", func() {
		It("fails on assertion that is not base64 encoded", func() {
			body := &appvalidate.AssertionVerify{
				KeyID: attestedVerifiedUser.KeyID,
				ClientData: appvalidate.AssertionClientData{
					Challenge: challenge,
				},
				Assertion: `{"key": "field"}`,
			}

			req := newRequest(http.MethodPost, "/v1/assertions/verifications", attestedVerifiedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
		It("fails on incorrect assertion", func() {
			body := &appvalidate.AssertionVerify{
				KeyID: attestedVerifiedUser.KeyID,
				ClientData: appvalidate.AssertionClientData{
					Challenge: challenge,
				},
				Assertion: `YWJjZGVm`,
			}

			req := newRequest(http.MethodPost, "/v1/assertions/verifications", attestedVerifiedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})

// user is a helper user that contains relevant user information for tests.
type user struct {
	UserID               string
	SessionToken         string
	Details              request.AuthDetails
	KeyID                string
	AttestationVerified  bool
	AttestationChallenge string
	AssertionChallenge   string
}

// newRequest wraps httptest.NewRequest w/ a default logger as some of the
// middleware expect the logger to be present so this prevents a nil pointer
// dereference. body can be nil, an io.Reader, or a struct that is assumed to
// be JSON marshalable
func newRequest(method, url, sessionToken string, body interface{}) *http.Request {
	var newBody io.Reader
	var contentType string

	if body != nil {
		switch v := body.(type) {
		case string:
			newBody = strings.NewReader(v)
		case []byte:
			newBody = bytes.NewReader(v)
		case io.Reader:
			newBody = v
		default:
			body, err := json.Marshal(v)
			if err == nil {
				newBody = bytes.NewReader(body)
				contentType = "application/json"
			}
		}
	}
	req := httptest.NewRequest(method, url, newBody)
	if contentType != "" {
		req.Header.Add("content-type", contentType)
	}
	if sessionToken != "" {
		req.Header.Add(auth.TidepoolSessionTokenHeaderKey, sessionToken)
	}
	ctx := log.NewContextWithLogger(req.Context(), logTest.NewLogger())
	return req.Clone(ctx)
}

func unmashalBody(r io.ReadCloser, result interface{}) error {
	defer r.Close()
	return json.NewDecoder(r).Decode(result)
}

func newRepository(ctrl *gomock.Controller, initialValidations []appvalidate.AppValidation) *appvalidate.MockRepository {
	// In memory map for persistence across calls.
	// [appvalidate.Filter] => *appvalidate.AppValidation
	mapping := &sync.Map{}

	for _, appValidation := range initialValidations {
		// Make a copy since storing &appValidation is shared in the range loop.
		copy := appValidation
		mapping.Store(appvalidate.Filter{UserID: copy.UserID, KeyID: copy.KeyID}, &copy)
	}

	repo := appvalidate.NewMockRepository(ctrl)
	repo.EXPECT().
		Upsert(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, v *appvalidate.AppValidation) error {
			mapping.Store(appvalidate.Filter{UserID: v.UserID, KeyID: v.KeyID}, v)
			return nil
		}).
		AnyTimes()

	repo.EXPECT().
		IsVerified(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f appvalidate.Filter) (bool, error) {
			verificationRaw, ok := mapping.Load(f)
			if !ok {
				return false, errors.New("not found")
			}
			return verificationRaw.(*appvalidate.AppValidation).Verified, nil
		}).
		AnyTimes()

	repo.EXPECT().
		GetAttestationChallenge(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f appvalidate.Filter) (string, error) {
			verificationRaw, ok := mapping.Load(f)
			if !ok {
				return "", errors.New("not found")
			}
			return verificationRaw.(*appvalidate.AppValidation).AttestationChallenge, nil
		}).
		AnyTimes()

	repo.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f appvalidate.Filter) (*appvalidate.AppValidation, error) {
			verificationRaw, ok := mapping.Load(f)
			if !ok {
				return nil, errors.New("not found")
			}
			return verificationRaw.(*appvalidate.AppValidation), nil
		}).
		AnyTimes()

	repo.EXPECT().
		UpdateAssertion(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f appvalidate.Filter, u appvalidate.AssertionUpdate) error {
			verificationRaw, ok := mapping.Load(f)
			if !ok {
				return errors.New("not found")
			}
			verification := verificationRaw.(*appvalidate.AppValidation)
			// Ignore zero values like the `bson:",omitempty"` tag does
			if !u.VerifiedTime.IsZero() {
				verification.AssertionVerifiedTime = &u.VerifiedTime
			}
			if u.AssertionCounter > 0 {
				verification.AssertionCounter = u.AssertionCounter
			}
			if u.Challenge != "" {
				verification.AssertionChallenge = u.Challenge
			}
			return nil
		}).
		AnyTimes()

	repo.EXPECT().
		UpdateAttestation(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, f appvalidate.Filter, u appvalidate.AttestationUpdate) error {
			verificationRaw, ok := mapping.Load(f)
			if !ok {
				return errors.New("not found")
			}
			verification := verificationRaw.(*appvalidate.AppValidation)
			verification.PublicKey = u.PublicKey
			verification.Verified = u.Verified
			verification.FraudAssessmentReceipt = u.FraudAssessmentReceipt
			verification.AttestationVerifiedTime = &u.VerifiedTime
			return nil
		}).
		AnyTimes()

	return repo
}
