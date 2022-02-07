package clinics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	clinic "github.com/tidepool-org/clinic/client"
	"go.uber.org/fx"

	// workaround for mockgen incompatibility with vendored dependencies
	_ "github.com/golang/mock/mockgen/model"

	"github.com/tidepool-org/platform/auth"
)

var ClientModule = fx.Provide(NewClient)

//go:generate mockgen --build_flags=--mod=mod -source=./service.go -destination=./mock.go -package clinics Client

type Client interface {
	GetClinician(ctx context.Context, clinicID, clinicianID string) (*clinic.Clinician, error)
	SharePatientAccount(ctx context.Context, clinicID, patientID string) (*clinic.Patient, error)
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

		req.Header.Add(auth.TidepoolSessionTokenHeaderKey, token)
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
		return nil, fmt.Errorf("unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
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
		return nil, fmt.Errorf("unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
	}
	return response.JSON200, nil
}

func (d *defaultClient) getPatient(ctx context.Context, clinicID, patientID string) (*clinic.Patient, error) {
	response, err := d.httpClient.GetPatientWithResponse(ctx, clinic.ClinicId(clinicID), clinic.PatientId(patientID))
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
	}
	return response.JSON200, nil
}
