package device

import "github.com/tidepool-org/platform/structure"

const (
	DeviceTypePump = "pump"
	DeviceTypeCGM  = "cgm"
)

type Device struct {
	Type         string `json:"type"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
}

func (d *Device) Validate(validator structure.Validator) {
	// TODO: Validate device against the list of all supported devices
}

func Types() []string {
	return []string{
		DeviceTypePump,
		DeviceTypeCGM,
	}
}

func GetSupportedPumps() []Device {
	// TODO: Create a list with all supported pumps
	return []Device{}
}

func GetSupportedCGMs() []Device {
	// TODO: Create a list with all supported CGMs
	return []Device{}
}
