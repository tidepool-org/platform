package clinics

import (
	"context"
	"net/http"
	"net/url"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"

	clinic "github.com/tidepool-org/clinic/client"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
)

//go:generate mockgen -source=service.go -destination=test/service_mocks.go -package=test -typed

const ErrorCodeClinicClientFailure = "clinic-client-failure"

var ClientModule = fx.Provide(NewClient)

type Client interface {
	GetClinic(ctx context.Context, clinicID string) (*clinic.ClinicV1, error)
	GetClinician(ctx context.Context, clinicID, clinicianID string) (*clinic.ClinicianV1, error)
	GetEHRSettings(ctx context.Context, clinicId string) (*clinic.EhrSettingsV1, error)
	SharePatientAccount(ctx context.Context, clinicID, patientID string) (*clinic.PatientV1, error)
	ListEHREnabledClinics(ctx context.Context) ([]clinic.ClinicV1, error)
	SyncEHRData(ctx context.Context, clinicID string) error
	SyncEHRDataForPatient(ctx context.Context, patientID string) error
	GetPatients(ctx context.Context, clinicId string, userToken string, params *clinic.ListPatientsParams, injectedParams url.Values) ([]clinic.PatientV1, error)
	GetPatient(ctx context.Context, clinicID, patientID string) (*clinic.PatientV1, error)
	ListClinicsForPatient(ctx context.Context, patientID string) ([]clinic.PatientClinicRelationshipV1, error)
	UpdatePatientSummary(ctx context.Context, patientID string, glycemicRangesKey *string, patientSummary *clinic.PatientSummaryV1) (bool, error)
	// DeletePatientSummary deletes the summary with the id from every patient record
	// holding it.
	//
	// Deprecated: summaries computed for a glycemic range set have no id; use
	// [Client.DeletePatientSummaryOfType]. It is kept for work created before it existed.
	DeletePatientSummary(ctx context.Context, summaryID string) error
	DeletePatientSummaryOfType(ctx context.Context, patientID string, summaryType string) error
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

func (d *defaultClient) GetClinician(ctx context.Context, clinicID, clinicianID string) (*clinic.ClinicianV1, error) {
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

func (d *defaultClient) GetClinic(ctx context.Context, clinicID string) (*clinic.ClinicV1, error) {
	response, err := d.httpClient.GetClinicWithResponse(ctx, clinic.ClinicId(clinicID))
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

func (d *defaultClient) ListEHREnabledClinics(ctx context.Context) ([]clinic.ClinicV1, error) {
	offset := 0
	batchSize := 1000

	clinics := make([]clinic.ClinicV1, 0)
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

func (d *defaultClient) GetEHRSettings(ctx context.Context, clinicId string) (*clinic.EhrSettingsV1, error) {
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

func (d *defaultClient) SharePatientAccount(ctx context.Context, clinicID, patientID string) (*clinic.PatientV1, error) {
	permission := make(map[string]interface{}, 0)
	body := clinic.CreatePatientFromUserJSONRequestBody{
		Permissions: &clinic.PatientPermissionsV1{
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
		return d.GetPatient(ctx, clinicID, patientID)
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

// SyncEHRDataForPatient reports no error when the clinic service reports the patient has no active
// subscription to any clinic enabled for an electronic health record, which it does as not found. Most
// users are not such a patient, so reporting that as a failure would fail the work of nearly every user.
func (d *defaultClient) SyncEHRDataForPatient(ctx context.Context, patientID string) error {
	response, err := d.httpClient.SyncEHRDataForPatientWithResponse(ctx, clinic.PatientId(patientID))
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusAccepted && response.StatusCode() != http.StatusNotFound {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return err
	}
	return nil
}

// ListClinicsForPatient returns the patient records of the user, each with its clinic,
// or none when the clinic service reports the user as not found.
func (d *defaultClient) ListClinicsForPatient(ctx context.Context, patientID string) ([]clinic.PatientClinicRelationshipV1, error) {
	return page.CollectWithSize(page.PaginationSizeMaximum,
		func(pagination page.Pagination) ([]clinic.PatientClinicRelationshipV1, error) {
			response, err := d.httpClient.ListClinicsForPatientWithResponse(ctx, patientID,
				&clinic.ListClinicsForPatientParams{
					Offset: pointer.FromInt(pagination.Page * pagination.Size),
					Limit:  pointer.FromInt(pagination.Size),
				})
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
			if response.JSON200 == nil {
				return nil, nil
			}
			return *response.JSON200, nil
		})
}

// UpdatePatientSummary reports false, without an error, when the clinic service answers
// that the user is not a patient of any clinic, as no change or not found. Most users are
// not, so treating that as a failure would fail the work of nearly every user.
func (d *defaultClient) UpdatePatientSummary(ctx context.Context, patientID string, glycemicRangesKey *string, patientSummary *clinic.PatientSummaryV1) (bool, error) {
	params := &clinic.UpdatePatientSummaryParams{GlycemicRangesKey: glycemicRangesKey}
	response, err := d.httpClient.UpdatePatientSummaryWithResponse(ctx, clinic.PatientId(patientID), params, *patientSummary)
	if err != nil {
		return false, err
	}
	if response.StatusCode() == http.StatusNoContent || response.StatusCode() == http.StatusNotFound {
		return false, nil
	}
	if response.StatusCode() != http.StatusOK {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return false, err
	}
	return true, nil
}

// DeletePatientSummary reports no error when the clinic service reports the summary as not
// found, so that a delete resent by retried work targeting an already deleted summary does
// not fail the work forever.
func (d *defaultClient) DeletePatientSummary(ctx context.Context, summaryID string) error {
	response, err := d.httpClient.DeletePatientSummaryWithResponse(ctx, clinic.SummaryId(summaryID))
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusNoContent && response.StatusCode() != http.StatusNotFound {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return err
	}
	return nil
}

// DeletePatientSummaryOfType deletes the stats of the summary type from every patient
// record of the user, whatever their glycemic ranges. It tolerates the clinic service
// answering that the user is not a patient of any clinic, which it does as no change, so
// that a delete resent by retried work does not fail the work forever. Not found is an
// error: the clinic service never answers so for this route, so it can only come from a
// clinic service deployed without the route, and the delete must be retried.
func (d *defaultClient) DeletePatientSummaryOfType(ctx context.Context, patientID string, summaryType string) error {
	response, err := d.httpClient.DeletePatientSummaryOfTypeWithResponse(ctx, clinic.PatientId(patientID), summaryType)
	if err != nil {
		return err
	}
	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusNoContent {
		err = errors.Preparedf(ErrorCodeClinicClientFailure,
			"Unexpected status code from clinic service",
			"unexpected response status code %v from %v", response.StatusCode(), response.HTTPResponse.Request.URL)
		err = errors.WithMeta(err, response.HTTPResponse)
		return err
	}
	return nil
}

func (d *defaultClient) GetPatient(ctx context.Context, clinicID, patientID string) (*clinic.PatientV1, error) {
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

func (d *defaultClient) GetPatients(ctx context.Context, clinicId string, userToken string, params *clinic.ListPatientsParams, injectedParams url.Values) ([]clinic.PatientV1, error) {
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
