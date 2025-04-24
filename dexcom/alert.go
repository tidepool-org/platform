package dexcom

import (
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	AlertsResponseRecordType    = "alert"
	AlertsResponseRecordVersion = "3.0"

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

func AlertStates() []string {
	return []string{
		AlertStateUnknown,
		AlertStateInactive,
		AlertStateActiveSnoozed,
		AlertStateActiveAlarming,
	}
}

type AlertsResponse struct {
	RecordType    *string `json:"recordType,omitempty"`
	RecordVersion *string `json:"recordVersion,omitempty"`
	UserID        *string `json:"userId,omitempty"`
	Records       *Alerts `json:"records,omitempty"`
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
	parser = parser.WithMeta(a)

	a.RecordType = parser.String("recordType")
	a.RecordVersion = parser.String("recordVersion")
	a.UserID = parser.String("userId")
	a.Records = ParseAlerts(parser.WithReferenceArrayParser("records"))
}

func (a *AlertsResponse) Validate(validator structure.Validator) {
	validator = validator.WithMeta(a)

	validator.String("recordType", a.RecordType).Exists().EqualTo(AlertsResponseRecordType)
	validator.String("recordVersion", a.RecordVersion).Exists().EqualTo(AlertsResponseRecordVersion)
	validator.String("userId", a.UserID).Exists().NotEmpty()

	// Only validate that the records exist, remaining validation will occur later on a per-record basis
	if a.Records == nil {
		validator.WithReference("records").ReportError(structureValidator.ErrorValueNotExists())
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

type Alert struct {
	RecordID              *string `json:"recordId,omitempty"`
	SystemTime            *Time   `json:"systemTime,omitempty"`
	DisplayTime           *Time   `json:"displayTime,omitempty"`
	AlertName             *string `json:"alertName,omitempty"`
	AlertState            *string `json:"alertState,omitempty"`
	TransmitterGeneration *string `json:"transmitterGeneration,omitempty"`
	TransmitterID         *string `json:"transmitterId,omitempty"`
	DisplayDevice         *string `json:"displayDevice,omitempty"`
	DisplayApp            *string `json:"displayApp,omitempty"`
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
	parser = parser.WithMeta(a)

	a.RecordID = parser.String("recordId")
	a.SystemTime = ParseTime(parser, "systemTime")
	a.DisplayTime = ParseTime(parser, "displayTime")
	a.AlertName = parser.String("alertName")
	a.AlertState = parser.String("alertState")
	a.TransmitterGeneration = parser.String("transmitterGeneration")
	a.TransmitterID = parser.String("transmitterId")
	a.DisplayDevice = parser.String("displayDevice")
	a.DisplayApp = parser.String("displayApp")
}

func (a *Alert) Validate(validator structure.Validator) {
	validator = validator.WithMeta(a)

	validator.String("recordId", a.RecordID).Exists().NotEmpty()
	validator.Time("systemTime", a.SystemTime.Raw()).Exists().NotZero()
	validator.Time("displayTime", a.DisplayTime.Raw()).Exists().NotZero()
	validator.String("alertName", a.AlertName).Exists().OneOf(AlertNames()...)
	validator.String("alertState", a.AlertState).Exists().OneOf(AlertStates()...)
	validator.String("transmitterGeneration", a.TransmitterGeneration).Exists().OneOf(DeviceTransmitterGenerations()...)
	validator.String("transmitterId", a.TransmitterID).Exists().Using(TransmitterIDValidator)
	validator.String("displayDevice", a.DisplayDevice).Exists().OneOf(DeviceDisplayDevices()...)
	validator.String("displayApp", a.DisplayApp).Exists().OneOf(DeviceDisplayApps()...)

	// Log various warnings
	logger := validator.Logger().WithField("meta", a)
	if a.AlertName != nil && *a.AlertName == AlertNameUnknown {
		logger.Warnf("AlertName is '%s'", *a.AlertName)
	}
	if a.AlertState != nil && *a.AlertState == AlertStateUnknown {
		logger.Warnf("AlertState is '%s'", *a.AlertState)
	}
	if a.TransmitterID != nil && *a.TransmitterID == "" {
		logger.Warnf("TransmitterID is empty", *a.TransmitterID)
	}
}

func (a *Alert) Normalize(normalizer structure.Normalizer) {}
