package reporters

import (
	"context"
	"errors"
	"fmt"
	"time"

	clinic "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	realtimeDaysThreshold         = 16
	realtimePatientsLengthLimit   = 1000
	realtimePatientsInsuranceCode = "CPT-99454"
)

type PatientRealtimeDaysReporter struct {
	summarizer summary.Summarizer[*types.ContinuousStats, types.ContinuousStats]
}

func NewReporter(registry *summary.SummarizerRegistry) *PatientRealtimeDaysReporter {
	summarizer := summary.GetSummarizer[*types.ContinuousStats](registry)
	return &PatientRealtimeDaysReporter{
		summarizer: summarizer,
	}
}

func (r *PatientRealtimeDaysReporter) GetRealtimeDaysForPatients(ctx context.Context, clinicsClient clinics.Client, clinicId string, token string, startTime time.Time, endTime time.Time, patientFilters *clinic.ListPatientsParams) (*PatientsRealtimeDaysResponse, error) {
	patientFilters.Limit = pointer.FromAny(realtimePatientsLengthLimit + 1)

	patients, err := clinicsClient.GetPatients(ctx, clinicId, token, patientFilters)
	if err != nil {
		return nil, err
	}

	if len(patients) > realtimePatientsLengthLimit {
		return nil, fmt.Errorf("too many patients in clinic for report to succeed. (%d > limit %d)", len(patients), realtimePatientsLengthLimit)
	}

	userIds := make([]string, len(patients))
	for i := 0; i < len(patients); i++ {
		userIds[i] = *patients[i].Id
	}

	userIdsRealtimeDays, err := r.GetRealtimeDaysForUsers(ctx, userIds, startTime, endTime)
	if err != nil {
		return nil, err
	}

	patientsResponse := make([]PatientRealtimeDaysResponse, len(userIdsRealtimeDays))
	var patient PatientRealtimeDaysResponse
	var sufficient bool

	// we want to put >= realtimeDaysThreshold records first in the list, before any <
	// so we will insert from both directions on the list, and meet in the middle
	frontIndex := 0
	rearIndex := len(userIdsRealtimeDays) - 1

	for i := 0; i < len(userIdsRealtimeDays); i++ {
		sufficient = userIdsRealtimeDays[*patients[i].Id] >= realtimeDaysThreshold
		patient = PatientRealtimeDaysResponse{
			Id:                *patients[i].Id,
			FullName:          patients[i].FullName,
			BirthDate:         patients[i].BirthDate.Format(time.DateOnly),
			MRN:               patients[i].Mrn,
			RealtimeDays:      userIdsRealtimeDays[*patients[i].Id],
			HasSufficientData: sufficient,
		}

		if sufficient {
			patientsResponse[frontIndex] = patient
			frontIndex++
		} else {
			patientsResponse[rearIndex] = patient
			rearIndex--
		}
	}

	return &PatientsRealtimeDaysResponse{
		Config: PatientsRealtimeDaysConfigResponse{
			Code:      realtimePatientsInsuranceCode,
			ClinicId:  clinicId,
			StartDate: startTime,
			EndDate:   endTime,
		},
		Results: patientsResponse,
	}, nil
}

func (r *PatientRealtimeDaysReporter) GetRealtimeDaysForUsers(ctx context.Context, userIds []string, startTime time.Time, endTime time.Time) (map[string]int, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userIds == nil {
		return nil, errors.New("userIds is missing")
	}
	if len(userIds) == 0 {
		return nil, errors.New("no userIds provided")
	}
	if startTime.IsZero() {
		return nil, errors.New("startTime is missing")
	}
	if endTime.IsZero() {
		return nil, errors.New("startTime is missing")
	}

	if startTime.After(endTime) {
		return nil, errors.New("startTime is after endTime")
	}

	endOfHour := time.Now().Truncate(time.Hour).Add(time.Second * 3599)
	if startTime.Before(endOfHour.AddDate(0, 0, -60)) {
		return nil, errors.New("startTime is too old ( >60d ago ) ")
	}

	realtimeUsers := make(map[string]int)

	for _, userId := range userIds {
		userSummary, err := r.summarizer.GetSummary(ctx, userId)
		if err != nil {
			return nil, err
		}

		if userSummary != nil && userSummary.Stats != nil {
			realtimeUsers[userId] = userSummary.Stats.GetNumberOfDaysWithRealtimeData(startTime, endTime)
		} else {
			realtimeUsers[userId] = 0
		}
	}

	return realtimeUsers, nil
}

type PatientsRealtimeDaysResponse struct {
	Config  PatientsRealtimeDaysConfigResponse `json:"config"`
	Results []PatientRealtimeDaysResponse      `json:"results"`
}

type PatientsRealtimeDaysConfigResponse struct {
	Code      string    `json:"code"`
	ClinicId  string    `json:"clinicId"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type PatientRealtimeDaysResponse struct {
	Id                string  `json:"id"`
	FullName          string  `json:"fullName"`
	BirthDate         string  `json:"birthDate"`
	MRN               *string `json:"mrn"`
	RealtimeDays      int     `json:"realtimeDays"`
	HasSufficientData bool    `json:"hasSufficientData"`
}

type PatientRealtimeDaysFilter struct {
	StartTime      *time.Time
	EndTime        *time.Time
	PatientFilters *clinic.ListPatientsParams
}

func NewPatientRealtimeDaysFilter() *PatientRealtimeDaysFilter {
	return &PatientRealtimeDaysFilter{}
}

func (d *PatientRealtimeDaysFilter) Parse(parser structure.ObjectParser) {
	d.StartTime = parser.Time("startDate", time.RFC3339)
	d.EndTime = parser.Time("endDate", time.RFC3339)

	d.PatientFilters = &clinic.ListPatientsParams{}
	parser.JSON("patientFilters", d.PatientFilters)
}

func (d *PatientRealtimeDaysFilter) Validate(validator structure.Validator) {
	validator.Time("startDate", d.StartTime).NotZero()
	validator.Time("endDate", d.EndTime).NotZero()
}
