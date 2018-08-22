package dexcom

import (
	"strconv"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type DevicesResponse struct {
	Devices *Devices `json:"devices,omitempty"`
}

func ParseDevicesResponse(parser structure.ObjectParser) *DevicesResponse {
	if !parser.Exists() {
		return nil
	}
	datum := NewDevicesResponse()
	parser.Parse(datum)
	return datum
}

func NewDevicesResponse() *DevicesResponse {
	return &DevicesResponse{}
}

func (d *DevicesResponse) Parse(parser structure.ObjectParser) {
	d.Devices = ParseDevices(parser.WithReferenceArrayParser("devices"))
}

func (d *DevicesResponse) Validate(validator structure.Validator) {
	if devicesValidator := validator.WithReference("devices"); d.Devices != nil {
		devicesValidator.Validate(d.Devices)
	} else {
		devicesValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

type Devices []*Device

func ParseDevices(parser structure.ArrayParser) *Devices {
	if !parser.Exists() {
		return nil
	}
	datum := NewDevices()
	parser.Parse(datum)
	return datum
}

func NewDevices() *Devices {
	return &Devices{}
}

func (d *Devices) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*d = append(*d, ParseDevice(parser.WithReferenceObjectParser(reference)))
	}
}

func (d *Devices) Validate(validator structure.Validator) {
	for index, device := range *d {
		if deviceValidator := validator.WithReference(strconv.Itoa(index)); device != nil {
			deviceValidator.Validate(device)
		} else {
			deviceValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

type Device struct {
	Model             *string        `json:"model,omitempty"`
	LastUploadDate    *time.Time     `json:"lastUploadDate,omitempty"`
	AlertSettings     *AlertSettings `json:"alertSettings,omitempty"`
	UDI               *string        `json:"udi,omitempty"`
	SerialNumber      *string        `json:"serialNumber,omitempty"`
	TransmitterID     *string        `json:"transmitterId,omitempty"`
	SoftwareVersion   *string        `json:"softwareVersion,omitempty"`
	SoftwareNumber    *string        `json:"softwareNumber,omitempty"`
	Language          *string        `json:"language,omitempty"`
	IsMmolDisplayMode *bool          `json:"isMmolDisplayMode,omitempty"`
	IsBlindedMode     *bool          `json:"isBlindedMode,omitempty"`
	Is24HourMode      *bool          `json:"is24HourMode,omitempty"`
	DisplayTimeOffset *int           `json:"displayTimeOffset,omitempty"`
	SystemTimeOffset  *int           `json:"systemTimeOffset,omitempty"`
}

func ParseDevice(parser structure.ObjectParser) *Device {
	if !parser.Exists() {
		return nil
	}
	datum := NewDevice()
	parser.Parse(datum)
	return datum
}

func NewDevice() *Device {
	return &Device{}
}

func (d *Device) Parse(parser structure.ObjectParser) {
	d.Model = parser.String("model")
	d.LastUploadDate = parser.Time("lastUploadDate", DateTimeFormat)
	d.AlertSettings = ParseAlertSettings(parser.WithReferenceArrayParser("alertSettings"))
	d.UDI = parser.String("udi")
	d.SerialNumber = parser.String("serialNumber")
	d.TransmitterID = parser.String("transmitterId")
	d.SoftwareVersion = parser.String("softwareVersion")
	d.SoftwareNumber = parser.String("softwareNumber")
	d.Language = parser.String("language")
	d.IsMmolDisplayMode = parser.Bool("isMmolDisplayMode")
	d.IsBlindedMode = parser.Bool("isBlindedMode")
	d.Is24HourMode = parser.Bool("is24HourMode")
	d.DisplayTimeOffset = parser.Int("displayTimeOffset")
	d.SystemTimeOffset = parser.Int("systemTimeOffset")
}

func (d *Device) Validate(validator structure.Validator) {
	validator = validator.WithMeta(d)
	validator.String("model", d.Model).Exists().OneOf(ModelG5MobileApp, ModelG5Receiver, ModelG4WithShareReceiver, ModelG4Receiver, ModelUnknown)
	validator.Time("lastUploadDate", d.LastUploadDate).Exists().NotZero()
	if alertSettingsValidator := validator.WithReference("alertSettings"); d.AlertSettings != nil {
		alertSettingsValidator.Validate(d.AlertSettings)
	} else {
		alertSettingsValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("udi", d.UDI).NotEmpty()
	validator.String("serialNumber", d.SerialNumber).NotEmpty()
	validator.String("transmitterId", d.TransmitterID).Matches(transmitterIDExpression)
	validator.String("softwareVersion", d.SoftwareVersion).NotEmpty()
	validator.String("softwareNumber", d.SoftwareNumber).NotEmpty()
	validator.String("language", d.Language).NotEmpty()
}

type AlertSettings []*AlertSetting

func ParseAlertSettings(parser structure.ArrayParser) *AlertSettings {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertSettings()
	parser.Parse(datum)
	return datum
}

func NewAlertSettings() *AlertSettings {
	return &AlertSettings{}
}

func (a *AlertSettings) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*a = append(*a, ParseAlertSetting(parser.WithReferenceObjectParser(reference)))
	}
}

func (a *AlertSettings) Validate(validator structure.Validator) {
	for index, alertSetting := range *a {
		if alertSettingValidator := validator.WithReference(strconv.Itoa(index)); alertSetting != nil {
			alertSettingValidator.Validate(alertSetting)
		} else {
			alertSettingValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}

	if !validator.HasError() {
		a.Deduplicate() // HACK: Dexcom - duplicates not allowed; delete older
	}
}

func (a *AlertSettings) Deduplicate() {
	alertSettings := []*AlertSetting{}
	alertNameMap := map[string]bool{}
	for outerIndex, alertSetting := range *a {
		if alertSetting.AlertName == nil {
			alertSettings = append(alertSettings, alertSetting)
		} else if !alertNameMap[*alertSetting.AlertName] {
			for innerIndex := outerIndex + 1; innerIndex < len(*a); innerIndex++ {
				if (*a)[innerIndex].IsNewerMatchThan(alertSetting) {
					alertSetting = nil
					break
				}
			}
			if alertSetting != nil {
				alertSettings = append(alertSettings, alertSetting)
				alertNameMap[*alertSetting.AlertName] = true
			}
		}
	}
	*a = alertSettings
}

type AlertSetting struct {
	SystemTime  *time.Time `json:"systemTime,omitempty"`
	DisplayTime *time.Time `json:"displayTime,omitempty"`
	AlertName   *string    `json:"alertName,omitempty"`
	Unit        *string    `json:"unit,omitempty"`
	Value       *float64   `json:"value,omitempty"`
	Delay       *int       `json:"delay,omitempty"`
	Snooze      *int       `json:"snooze,omitempty"`
	Enabled     *bool      `json:"enabled,omitempty"`
}

func NewAlertSetting() *AlertSetting {
	return &AlertSetting{}
}

func ParseAlertSetting(parser structure.ObjectParser) *AlertSetting {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertSetting()
	parser.Parse(datum)
	return datum
}

func (a *AlertSetting) Parse(parser structure.ObjectParser) {
	a.SystemTime = parser.Time("systemTime", DateTimeFormat)
	a.DisplayTime = parser.Time("displayTime", DateTimeFormat)
	a.AlertName = parser.String("alertName")
	a.Unit = parser.String("unit")
	a.Value = parser.Float64("value")
	a.Delay = parser.Int("delay")
	a.Snooze = parser.Int("snooze")
	a.Enabled = parser.Bool("enabled")
}

func (a *AlertSetting) Validate(validator structure.Validator) {
	validator = validator.WithMeta(a)
	validator.Time("systemTime", a.SystemTime).Exists().NotZero().BeforeNow(NowThreshold)
	validator.Time("displayTime", a.DisplayTime).Exists().NotZero()
	validator.String("alertName", a.AlertName).Exists().OneOf(AlertNames()...)

	if a.AlertName != nil {
		switch *a.AlertName {
		case AlertNameFixedLow:
			a.validateFixedLow(validator)
		case AlertNameLow:
			a.validateLow(validator)
		case AlertNameHigh:
			a.validateHigh(validator)
		case AlertNameRise:
			a.validateRise(validator)
		case AlertNameFall:
			a.validateFall(validator)
		case AlertNameOutOfRange:
			a.validateOutOfRange(validator)
		}
	}
}

func (a *AlertSetting) IsNewerMatchThan(alertSetting *AlertSetting) bool {
	return a.AlertName != nil && alertSetting.AlertName != nil && *a.AlertName == *alertSetting.AlertName &&
		a.SystemTime != nil && alertSetting.SystemTime != nil && a.SystemTime.After(*alertSetting.SystemTime)
}

func (a *AlertSetting) validateFixedLow(validator structure.Validator) {
	// HACK: Dexcom - snooze of 28 is invalid; use snooze of 30 instead (per Dexcom)
	if a.Snooze != nil && *a.Snooze == 28 {
		*a.Snooze = 30
	}

	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingFixedLowUnits()...)
	if values := AlertSettingFixedLowValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", a.Value).Exists().OneOf(values...)
	}
	validator.Int("delay", a.Delay).OneOf(0)
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingFixedLowSnoozes()...)
	validator.Bool("enabled", a.Enabled).Exists().True()
}

func (a *AlertSetting) validateLow(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingLowUnits()...)
	if values := AlertSettingLowValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", a.Value).Exists().OneOf(values...)
	}
	validator.Int("delay", a.Delay).OneOf(0)
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingLowSnoozes()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateHigh(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingHighUnits()...)
	if values := AlertSettingHighValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", a.Value).Exists().OneOf(values...)
	}
	validator.Int("delay", a.Delay).OneOf(0)
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingHighSnoozes()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateRise(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingRiseUnits()...)
	if values := AlertSettingRiseValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", a.Value).Exists().OneOf(values...)
	}
	validator.Int("delay", a.Delay).OneOf(0)
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingRiseSnoozes()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateFall(validator structure.Validator) {
	// HACK: Dexcom - negative value is invalid; use positive value instead (per Dexcom)
	if a.Value != nil && *a.Value < 0 {
		*a.Value = -*a.Value
	}

	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingFallUnits()...)
	if values := AlertSettingFallValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", a.Value).Exists().OneOf(values...)
	}
	validator.Int("delay", a.Delay).OneOf(0)
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingFallSnoozes()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func (a *AlertSetting) validateOutOfRange(validator structure.Validator) {
	validator.String("unit", a.Unit).Exists().OneOf(AlertSettingOutOfRangeUnits()...)
	if values := AlertSettingOutOfRangeValuesForUnits(a.Unit); values != nil {
		validator.Float64("value", a.Value).Exists().OneOf(values...)
	}
	validator.Int("delay", a.Delay).OneOf(0)
	validator.Int("snooze", a.Snooze).Exists().OneOf(AlertSettingOutOfRangeSnoozes()...)
	validator.Bool("enabled", a.Enabled).Exists()
}

func AlertSettingFixedLowUnits() []string {
	return []string{UnitMgdL} // TODO: Add UnitMmolL
}

func AlertSettingFixedLowValuesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case UnitMgdL:
			return []float64{55}
		case UnitMmolL:
			return nil // TODO: Add values
		}
	}
	return nil
}

func AlertSettingFixedLowSnoozes() []int {
	return []int{0, 30}
}

func AlertSettingLowUnits() []string {
	return []string{UnitMgdL} // TODO: Add UnitMmolL
}

func AlertSettingLowValuesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case UnitMgdL:
			return alertSettingLowValuesMgdL
		case UnitMmolL:
			return alertSettingLowValuesMmolL
		}
	}
	return nil
}

var alertSettingLowValuesMgdL = generateFloatRange(60, 100, 5)
var alertSettingLowValuesMmolL []float64 // TODO: Add values

func AlertSettingLowSnoozes() []int {
	return alertSettingLowSnoozes
}

var alertSettingLowSnoozes = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)

func AlertSettingHighUnits() []string {
	return []string{UnitMgdL} // TODO: Add UnitMmolL
}

func AlertSettingHighValuesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case UnitMgdL:
			return alertSettingHighValuesMgdL
		case UnitMmolL:
			return alertSettingHighValuesMmolL
		}
	}
	return nil
}

var alertSettingHighValuesMgdL = generateFloatRange(120, 400, 10)
var alertSettingHighValuesMmolL []float64 // TODO: Add values

func AlertSettingHighSnoozes() []int {
	return alertSettingsHighSnoozes
}

var alertSettingsHighSnoozes = append(append([]int{0}, generateIntegerRange(15, 240, 5)...), generateIntegerRange(255, 300, 15)...)

func AlertSettingRiseUnits() []string {
	return []string{UnitMgdLMin} // TODO: UnitMmolLMin
}

func AlertSettingRiseValuesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case UnitMgdLMin:
			return []float64{2, 3}
		case UnitMmolLMin:
			return nil // TODO: Add values
		}
	}
	return nil
}

func AlertSettingRiseSnoozes() []int {
	return []int{0, 30}
}

func AlertSettingFallUnits() []string {
	return []string{UnitMgdLMin} // TODO: UnitMmolLMin
}

func AlertSettingFallValuesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case UnitMgdLMin:
			return []float64{2, 3}
		case UnitMmolLMin:
			return nil // TODO: Add values
		}
	}
	return nil
}

func AlertSettingFallSnoozes() []int {
	return []int{0, 30}
}

func AlertSettingOutOfRangeUnits() []string {
	return []string{UnitMinutes}
}

func AlertSettingOutOfRangeValuesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case UnitMinutes:
			return alertSettingOutOfRangeValuesMinutes
		}
	}
	return nil
}

var alertSettingOutOfRangeValuesMinutes = generateFloatRange(20, 240, 5)

func AlertSettingOutOfRangeSnoozes() []int {
	return []int{0, 20, 25, 30}
}

func generateFloatRange(min float64, max float64, step float64) []float64 {
	r := []float64{}
	for v := min; v <= max; v += step {
		r = append(r, v)
	}
	return r
}

func generateIntegerRange(min int, max int, step int) []int {
	r := []int{}
	for v := min; v <= max; v += step {
		r = append(r, v)
	}
	return r
}
