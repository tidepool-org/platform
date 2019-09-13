package status

import "github.com/tidepool-org/platform/structure"

type AlertsArray *[]string

type AlertsStruct struct {
	Alerts *[]string `json:"alerts,omitempty" bson:"alerts,omitempty"`
}

func (a *AlertsStruct) statusObject() {
}

func ParseAlertsStruct(parser structure.ObjectParser) *AlertsStruct {
	if !parser.Exists() {
		return nil
	}
	datum := NewAlertsStruct()
	parser.Parse(datum)
	return datum
}

func NewAlertsStruct() *AlertsStruct {
	return &AlertsStruct{}
}

func (a *AlertsStruct) Parse(parser structure.ObjectParser) {
	a.Alerts = parser.StringArray("alerts")
}
