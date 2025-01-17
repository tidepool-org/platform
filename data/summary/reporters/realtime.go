package reporters

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/types"
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

func (r *PatientRealtimeDaysReporter) GetRealtimeDaysForPatients(ctx context.Context, clinicsClient clinics.Client, clinicId string, token string, startTime time.Time, endTime time.Time, patientFilters map[string]any) (*PatientsRealtimeDaysResponse, error) {
	injectedParams := map[string][]string{}
	for p, v := range patientFilters {
		var finalParam []string

		// handle tags array specifically, as it doesn't convert direct from json
		if p == "tags" {
			tagsAny := v.([]any)
			finalParam = make([]string, len(tagsAny))
			for i := range tagsAny {
				finalParam[i] = fmt.Sprintf("%v", tagsAny[i].(string))
			}

			finalParam = []string{strings.Join(finalParam, ",")}
		} else {
			finalParam = []string{fmt.Sprintf("%v", v)}
		}

		injectedParams[p] = finalParam
	}

	injectedParams["limit"] = []string{fmt.Sprintf("%d", realtimePatientsLengthLimit+1)}

	patients, err := clinicsClient.GetPatients(ctx, clinicId, token, nil, injectedParams)
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
	PatientFilters map[string]any
}

func NewPatientRealtimeDaysFilter() *PatientRealtimeDaysFilter {
	return &PatientRealtimeDaysFilter{}
}

func (d *PatientRealtimeDaysFilter) Parse(parser structure.ObjectParser) {
	d.StartTime = parser.Time("startDate", time.RFC3339)
	d.EndTime = parser.Time("endDate", time.RFC3339)

	d.PatientFilters = map[string]any{}
	parser.JSON("patientFilters", &d.PatientFilters)
}

func (d *PatientRealtimeDaysFilter) Validate(validator structure.Validator) {
	validator.Time("startDate", d.StartTime).NotZero()
	validator.Time("endDate", d.EndTime).NotZero()
}
