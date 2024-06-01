package dexcom

import (
	"strconv"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	AlertNameUnknown       = "unknown"
	AlertNameHigh          = "high"
	AlertNameLow           = "low"
	AlertNameRise          = "rise"
	AlertNameFall          = "fall"
	AlertNameOutOfRange    = "outOfRange"
	AlertNameUrgentLow     = "urgentLow"
	AlertNameUrgentLowSoon = "urgentLowSoon"
	AlertNameNoReadings    = "noReadings"
	AlertNameFixedLow      = "fixedLow"

	AlertStateUnknown        = "unknown"
	AlertStateInactive       = "inactive"
	AlertStateActiveSnoozed  = "activeSnoozed"
	AlertStateActiveAlarming = "activeAlarming"
)

func AlertStates() []string {
	return []string{
		AlertStateUnknown,
		AlertStateInactive,
		AlertStateActiveSnoozed,
		AlertStateActiveAlarming,
	}
}

func AlertNames() []string {
	return []string{
		AlertNameUnknown,
		AlertNameHigh,
		AlertNameLow,
		AlertNameRise,
		AlertNameFall,
		AlertNameOutOfRange,
		AlertNameUrgentLow,
		AlertNameUrgentLowSoon,
		AlertNameNoReadings,
		AlertNameFixedLow,
	}
}

type AlertsResponse struct {
	RecordType    *string `json:"recordType,omitempty"`
	RecordVersion *string `json:"recordVersion,omitempty"`
	UserID        *string `json:"userId,omitempty"`
	Alerts        *Alerts `json:"records,omitempty"`
}

func ParseAlertsResponse(parser structure.ObjectParser) *AlertsResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertsResponse()
	parser.Parse(datum)
	return datum
}

func NewAlertsResponse() *AlertsResponse {
	return &AlertsResponse{}
}

func (a *AlertsResponse) Parse(parser structure.ObjectParser) {
	a.UserID = parser.String("userId")
	a.RecordType = parser.String("recordType")
	a.RecordVersion = parser.String("recordVersion")
	a.Alerts = ParseAlerts(parser.WithReferenceArrayParser("records"))
}

func (a *AlertsResponse) Validate(validator structure.Validator) {
	if alertsValidator := validator.WithReference("records"); a.Alerts != nil {
		a.Alerts.Validate(alertsValidator)
	} else {
		alertsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type Alerts []*Alert

func ParseAlerts(parser structure.ArrayParser) *Alerts {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlerts()
	parser.Parse(datum)
	return datum
}

func NewAlerts() *Alerts {
	return &Alerts{}
}

func (a *Alerts) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*a = append(*a, ParseAlert(parser.WithReferenceObjectParser(reference)))
	}
}

func (a *Alerts) Validate(validator structure.Validator) {
	for index, egv := range *a {
		if alertValidator := validator.WithReference(strconv.Itoa(index)); egv != nil {
			egv.Validate(alertValidator)
		} else {
			alertValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Alert struct {
	AlertName  *string `json:"alertName,omitempty"`
	AlertState *string `json:"alertState,omitempty"`

	ID                    *string `json:"recordId,omitempty"`
	SystemTime            *Time   `json:"systemTime,omitempty"`
	DisplayTime           *Time   `json:"displayTime,omitempty"`
	TransmitterID         *string `json:"transmitterId,omitempty"`
	TransmitterGeneration *string `json:"transmitterGeneration,omitempty"`
	DisplayDevice         *string `json:"displayDevice,omitempty"`
}

func ParseAlert(parser structure.ObjectParser) *Alert {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlert()
	parser.Parse(datum)
	return datum
}

func NewAlert() *Alert {
	return &Alert{}
}

func (a *Alert) Parse(parser structure.ObjectParser) {
	a.ID = parser.String("recordId")
	a.SystemTime = ParseTime(parser, "systemTime")
	a.DisplayTime = ParseTime(parser, "displayTime")
	a.TransmitterID = parser.String("transmitterId")
	a.TransmitterGeneration = parser.String("transmitterGeneration")
	a.DisplayDevice = parser.String("displayDevice")
	a.AlertName = parser.String("alertName")
	a.AlertState = parser.String("alertState")
}

func (a *Alert) Validate(validator structure.Validator) {
	validator = validator.WithMeta(a)
	validator.String("recordId", a.ID).Exists().NotEmpty()
	validator.Time("systemTime", a.SystemTime.Raw()).NotZero().BeforeNow(SystemTimeNowThreshold)
	validator.Time("displayTime", a.DisplayTime.Raw()).NotZero()
	validator.String("alertName", a.AlertName).Exists().OneOf(AlertNames()...)
	validator.String("alertState", a.AlertState).Exists().OneOf(AlertStates()...)
	validator.String("transmitterGeneration", a.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("displayDevice", a.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)
	validator.String("transmitterId", a.TransmitterID).Using(TransmitterIDValidator)
}
