package device

const (
	DeviceTypePump = "pump"
	DeviceTypeCGM = "cgm"
)

type Device struct {
	Type		 string `json:"type"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
}

func DeviceTypes() []string {
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

