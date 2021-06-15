package clinics

import (
	"strings"

	api "github.com/tidepool-org/clinic/client"
)

func IsPrescriber(clinician *api.Clinician) bool {
	if clinician == nil {
		return false
	}
	for _, role := range clinician.Roles {
		if strings.ToLower(role) == "prescriber" {
			return true
		}
	}
	return false
}
