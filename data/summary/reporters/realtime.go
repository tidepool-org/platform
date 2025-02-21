package reporters

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

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
	summarizer summary.Summarizer[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
}

func NewReporter(registry *summary.SummarizerRegistry) *PatientRealtimeDaysReporter {
	summarizer := summary.GetSummarizer[*types.ContinuousPeriods, *types.ContinuousBucket](registry)
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

func (r *PatientRealtimeDaysReporter) GetNumberOfDaysWithRealtimeData(ctx context.Context, buckets *mongo.Cursor) (count int, err error) {
	tomorrow := time.Time{}
	previousBucketTime := time.Time{}

	for buckets.Next(ctx) {
		bucket := types.Bucket[*types.ContinuousBucket, types.ContinuousBucket]{}
		if err = buckets.Decode(&bucket); err != nil {
			return 0, err
		}

		if !previousBucketTime.IsZero() && bucket.Time.Compare(previousBucketTime) >= 0 {
			return 0, fmt.Errorf("bucket with date %s is equal or later than to the last added bucket with date %s, "+
				"buckets must be in reverse order and unique", bucket.Time, previousBucketTime)
		}
		previousBucketTime = bucket.Time

		if tomorrow.IsZero() {
			tomorrow = bucket.Time
		}

		// if before or equal to nextDay
		if bucket.Time.Compare(tomorrow) <= 0 && bucket.Data.Realtime.Records > 0 {
			count += 1

			// set tomorrow, but in the original hour/minute/second as the first bucket for day counting
			tomorrow = time.Date(bucket.Time.Year(), bucket.Time.Month(), bucket.Time.Day()-1,
				tomorrow.Hour(), tomorrow.Minute(), tomorrow.Second(), tomorrow.Nanosecond(), tomorrow.Location())
		}
	}

	return count, nil
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
		return nil, errors.New("endTime is missing")
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
		buckets, err := r.summarizer.GetBucketsRange(ctx, userId, startTime, endTime)
		if err != nil {
			return nil, err
		}

		realtimeUsers[userId], err = r.GetNumberOfDaysWithRealtimeData(ctx, buckets)
		if err != nil {
			return nil, err
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
