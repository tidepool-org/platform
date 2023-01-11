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
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/middleware"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -build_flags=--mod=mod -destination=./auth_service_mock.go -package=v1 -mock_names Service=MockAuthService github.com/tidepool-org/platform/auth/service Service

var _ = Describe("App Validation", func() {
	defer GinkgoRecover()

	// Note the setup is outside a BeforeEach because this would simulate
	// multiple calls to a single http.Handler which is more representative
	// of actual use as opposed to creating an http.Handler for every request.
	ctrl := gomock.NewController(GinkgoT())

	service := v1.NewMockAuthService(ctrl)
	service.EXPECT().
		Logger().
		Return(logTest.NewLogger()).
		AnyTimes()

	unattestedUser := user{
		UserID:              "unattested",
		SessionToken:        "unattestedSessionToken",
		Details:             request.NewDetails(request.MethodSessionToken, "unattested", "unattestedSessionToken"),
		AttestationVerified: false,
	}
	attestedUser := user{
		UserID:              "attested",
		SessionToken:        "attestedSessionToken",
		Details:             request.NewDetails(request.MethodSessionToken, "attested", "attestedSessionToken"),
		KeyID:               "YWJjZGVmYWJjZGVm",
		AttestationVerified: false,
	}
	attestedVerifiedUser := user{
		UserID:              "attestedVerified",
		SessionToken:        "attestedVerifiedSessionToken",
		Details:             request.NewDetails(request.MethodSessionToken, "attestedVerified", "attestedVerifiedSessionToken"),
		KeyID:               "YWJkZmRlZg=",
		AttestationVerified: true,
	}
	users := []user{
		unattestedUser,
		attestedUser,
		attestedVerifiedUser,
	}

	challenge := "challenge"
	serverSessionToken := "serverSessionToken"

	initialValidations := make([]appvalidate.AppValidation, len(users))
	for i, user := range users {
		validation := appvalidate.AppValidation{
			UserID:               user.UserID,
			KeyID:                user.KeyID,
			Verified:             user.AttestationVerified,
			AttestationChallenge: challenge,
		}
		if user.AttestationVerified {
			validation.AttestationVerifiedTime = pointer.FromTime(time.Date(2023, time.January, 3, 10, 0, 0, 0, time.UTC))
		}
		initialValidations[i] = validation
	}
	repo := newRepository(ctrl, initialValidations)
	generator := appvalidate.NewMockChallengeGenerator(ctrl)
	generator.EXPECT().
		GenerateChallenge(gomock.Any()).
		Return(challenge, nil).
		AnyTimes()
	validator, err := appvalidate.NewValidator(repo, generator, appvalidate.ValidatorConfig{
		AppleAppID:    "org.tidepool.app",
		ChallengeSize: 10,
	})
	Expect(err).ToNot(HaveOccurred())

	service.EXPECT().
		AppValidator().
		Return(validator).
		AnyTimes()

	authClient := auth.NewMockAuthClient(ctrl)
	authClient.EXPECT().
		ServerSessionToken().
		Return(serverSessionToken, nil).
		AnyTimes()
	authClient.EXPECT().
		ValidateSessionToken(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, token string) (request.Details, error) {
			for _, user := range users {
				if token == user.SessionToken {
					return user.Details, nil
				}
			}
			return nil, request.ErrorUnauthorized()
		}).
		AnyTimes()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, err := v1.NewRouter(service)
	Expect(err).ToNot(HaveOccurred())

	authMiddleware, err := middleware.NewAuth("secret", authClient)
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
	handler := api.MakeHandler()

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
				KeyID: "YWJjZGVmZ2hpamFiY2RlZmdoaWphYmNkZWZnaGlq",
			}

			req := newRequest(http.MethodPost, "/v1/assertions/challenges", unattestedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).ToNot(Equal(http.StatusCreated))
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
			Expect(w.Code).ToNot(Equal(http.StatusCreated))
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
		// Commented out tests right now because need an actual signed object
		// to verify - will get once available from actual iOS device.
		// It("succeeds on valid input", func() {
		// 	body := &appvalidate.AttestationVerify{
		// 		KeyID:             attestedUser.KeyID,
		// 		Challenge:         challenge,
		// 		AttestationObject: "YWJjZGVmZw==", // base64 encoded string of the binary CBOR data returned from iOS api.
		// 	}

		// 	req := newRequest(http.MethodPost, "/v1/attestations/verifications", attestedUser.SessionToken, body)
		// 	w := httptest.NewRecorder()
		// 	handler.ServeHTTP(w, req)
		// 	Expect(w.Code).To(Equal(http.StatusNoContent))
		// })

		It("fails on attestation object that is not base64 encoded", func() {
			body := &appvalidate.AttestationVerify{
				KeyID:             attestedUser.KeyID,
				Challenge:         challenge,
				AttestationObject: `{"key": "field"}`,
			}

			req := newRequest(http.MethodPost, "/v1/attestations/verifications", attestedUser.SessionToken, body)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})
	})
})

// user is a helper user that contains relevant user information for tests.
type user struct {
	UserID              string
	SessionToken        string
	Details             request.Details
	KeyID               string
	AttestationVerified bool
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
