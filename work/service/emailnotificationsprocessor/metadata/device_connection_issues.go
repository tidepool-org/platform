package metadata

import (
	"github.com/tidepool-org/platform/structure"
)

// DeviceConnectionIssuesData is the metadata added to a [work.Work] item for notifying users about a device connection issue, the fields will be filled out by clinic-worker when it responds to CDC events on DataSources that change the DataSource.State property.
type DeviceConnectionIssuesData struct {
	UserId            string
	ProviderName      string
	RestrictedTokenId string
	PatientName       string
	DataSourceState   string
	EmailTemplate     string
}

func (d *DeviceConnectionIssuesData) Parse(parser structure.ObjectParser) {
	d.UserId = *parser.String("userId")
	d.ProviderName = *parser.String("providerName")
	d.RestrictedTokenId = *parser.String("restrictedTokenId")
	d.PatientName = *parser.String("patientName")
	d.DataSourceState = *parser.String("dataSourceState")
	d.EmailTemplate = *parser.String("emailTemplate")
}

func (d *DeviceConnectionIssuesData) Validate(validator structure.Validator) {
	validator.String("userId", &d.UserId).NotEmpty()
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("restrictedTokenId", &d.RestrictedTokenId).NotEmpty()
	validator.String("patientName", &d.PatientName).NotEmpty()
	validator.String("dataSourceState", &d.DataSourceState).NotEmpty()
	validator.String("emailTemplate", &d.EmailTemplate).NotEmpty()
}
