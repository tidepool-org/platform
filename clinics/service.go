package clinics

import (
	"context"
	"net/http"
	"net/url"

	"github.com/tidepool-org/platform/errors"

	"github.com/tidepool-org/platform/pointer"

	"github.com/kelseyhightower/envconfig"
	clinic "github.com/tidepool-org/clinic/client"
	"go.uber.org/fx"

	"github.com/tidepool-org/platform/auth"
)

const ErrorCodeClinicClientFailure = "clinic-client-failure"

var ClientModule = fx.Provide(NewClient)

//go:generate mockgen -source=service.go -destination=test/service_mocks.go -package=test Client
type Client interface {
	GetClinician(ctx context.Context, clinicID, clinicianID string) (*clinic.Clinician, error)
	GetEHRSettings(ctx context.Context, clinicId string) (*clinic.EHRSettings, error)
	SharePatientAccount(ctx context.Context, clinicID, patientID string) (*clinic.Patient, error)
	ListEHREnabledClinics(ctx context.Context) ([]clinic.Clinic, error)
	SyncEHRData(ctx context.Context, clinicID string) error
	GetPatients(ctx context.Context, clinicId string, userToken string, params *clinic.ListPatientsParams, injectedParams url.Values) ([]clinic.Patient, error)
}

type config struct {
	ServiceAddress string `envconfig:"TIDEPOOL_CLINIC_CLIENT_ADDRESS"`
}

func (c *config) Load() error {
	return envconfig.Process("", c)
}

type defaultClient struct {
	httpClient clinic.ClientWithResponsesInterface
	authClient auth.ExternalAccessor
}

func NewClient(authClient auth.ExternalAccessor) (Client, error) {
	cfg := &config{}
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	opts := clinic.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		token, err := authClient.ServerSessionToken()
		if err != nil {
			return err
		}

		// conditionally set token only if not already present
		if req.Header.Get(auth.TidepoolSessionTokenHeaderKey) == "" {
			req.Header.Add(auth.TidepoolSessionTokenHeaderKey, token)
		}

		return nil
	})
	httpClient, err := clinic.NewClientWithResponses(cfg.ServiceAddress, opts)
	if err != nil {
		return nil, err
	}

	return &defaultClient{
		httpClient: httpClient,
		authClient: authClient,
	}, nil
}

func (d *defaultClient) GetClinician(ctx context.Context, clinicID, clinicianID string) (*clinic.Clinician, error) {
	response, err := d.httpClient.GetClinicianWithResponse(ctx, clinic.ClinicId(clinicID), clinic.ClinicianId(clinicianID))
	if err != nil {
		return nil, err
	}
	if response.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	if response.StatusCode() != http.StatusOK {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return nil, err
	}
	return response.JSON200, nil
}

func (d *defaultClient) ListEHREnabledClinics(ctx context.Context) ([]clinic.Clinic, error) {
	offset := 0
	batchSize := 1000

	clinics := make([]clinic.Clinic, 0)
	for {
		response, err := d.httpClient.ListClinicsWithResponse(ctx, &clinic.ListClinicsParams{
			EhrEnabled: pointer.FromBool(true),
			Offset:     &offset,
			Limit:      &batchSize,
		})
		if err != nil {
			return nil, err
		}
		if response.StatusCode() != http.StatusOK {
			err = errors.Preparedf(ErrorCodeClinicClientFailure,
				"Unexpected status code from clinic service",
				"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
			err = errors.WithMeta(err, response.HTTPResponse)
			return nil, err
		}
		if response.JSON200 == nil {
			break
		}

		clinics = append(clinics, *response.JSON200...)
		offset = offset + batchSize

		if len(*response.JSON200) < batchSize {
			break
		}
	}

	return clinics, nil
}

func (d *defaultClient) GetEHRSettings(ctx context.Context, clinicId string) (*clinic.EHRSettings, error) {
	response, err := d.httpClient.GetEHRSettingsWithResponse(ctx, clinicId)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || response.StatusCode() != http.StatusOK {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return nil, err
	}
	return response.JSON200, nil
}

func (d *defaultClient) SharePatientAccount(ctx context.Context, clinicID, patientID string) (*clinic.Patient, error) {
	permission := make(map[string]interface{}, 0)
	body := clinic.CreatePatientFromUserJSONRequestBody{
		Permissions: &clinic.PatientPermissions{
			Note: &permission,
			View: &permission,
		},
	}
	response, err := d.httpClient.CreatePatientFromUserWithResponse(ctx, clinic.ClinicId(clinicID), clinic.PatientId(patientID), body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() == http.StatusConflict {
		// User is already shared with the clinic
		return d.getPatient(ctx, clinicID, patientID)
	}
	if response.StatusCode() != http.StatusOK {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return nil, err
	}
	return response.JSON200, nil
}

func (d *defaultClient) SyncEHRData(ctx context.Context, clinicID string) error {
	response, err := d.httpClient.SyncEHRDataWithResponse(ctx, clinicID)
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusAccepted {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return err
	}
	return nil
}

func (d *defaultClient) getPatient(ctx context.Context, clinicID, patientID string) (*clinic.Patient, error) {
	response, err := d.httpClient.GetPatientWithResponse(ctx, clinic.ClinicId(clinicID), clinic.PatientId(patientID))
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return nil, err
	}
	return response.JSON200, nil
}

func (d *defaultClient) GetPatients(ctx context.Context, clinicId string, userToken string, params *clinic.ListPatientsParams, injectedParams url.Values) ([]clinic.Patient, error) {
	response, err := d.httpClient.ListPatientsWithResponse(ctx, clinicId, params, func(ctx context.Context, req *http.Request) error {
		if len(injectedParams) != 0 {
			q := req.URL.Query()
			for p, v := range injectedParams {
				for _, item := range v {
					q.Add(p, item)
				}
			}

			req.URL.RawQuery = q.Encode()
		}

		req.Header.Set(auth.TidepoolSessionTokenHeaderKey, userToken)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if response.StatusCode() != http.StatusOK {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)

		return nil, err
	}

	return *response.JSON200.Data, nil
}
