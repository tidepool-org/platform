package metadata

import (
	"github.com/tidepool-org/platform/structure"
)

// ConnectAccountReminderData is the metadata added to a [work.Work] item for reminding users to connect their account to a C2C provider.
type ConnectAccountReminderData struct {
	UserId            string
	PatientName       string
	ProviderName      string
	RestrictedTokenId string
	EmailTemplate     string
}

func (d *ConnectAccountReminderData) Parse(parser structure.ObjectParser) {
	d.UserId = *parser.String("userId")
	d.ProviderName = *parser.String("providerName")
	d.PatientName = *parser.String("patientName")
	d.RestrictedTokenId = *parser.String("restrictedTokenId")
	d.EmailTemplate = *parser.String("emailTemplate")
}

func (d ConnectAccountReminderData) Validate(validator structure.Validator) {
	validator.String("userId", &d.UserId).NotEmpty()
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("patientName", &d.PatientName).NotEmpty()
	validator.String("restrictedTokenId", &d.RestrictedTokenId).NotEmpty()
	validator.String("emailTemplate", &d.EmailTemplate).NotEmpty()
}
