package metadata

import (
	"github.com/tidepool-org/platform/structure"
)

// ClaimAccountReminderData is the metadata added to a [work.Work] item for reminding users to claim their account.
type ClaimAccountReminderData struct {
	UserId     string
	Email      string
	ClinicId   string
	ClinicName string
}

func (d *ClaimAccountReminderData) Parse(parser structure.ObjectParser) {
	d.UserId = *parser.String("userId")
	d.Email = *parser.String("email")
	d.ClinicId = *parser.String("clinicId")
	d.ClinicName = *parser.String("clinicName")
}

func (d *ClaimAccountReminderData) Validate(validator structure.Validator) {
	validator.String("userId", &d.UserId).NotEmpty()
	validator.String("email", &d.Email).NotEmpty()
	validator.String("clinicId", &d.ClinicId).NotEmpty()
	validator.String("clinicName", &d.ClinicName).NotEmpty()
}
