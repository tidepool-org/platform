package work

import (
	"maps"
	"slices"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeviceIDLengthMaximum     = 100
	DeviceHashLengthMaximum   = 100
	DeviceHashesLengthMaximum = 100
)

type DeviceHashes map[string]string

func ParseDeviceHashes(parser structure.ObjectParser) *DeviceHashes {
	if !parser.Exists() {
		return nil
	}
	datum := &DeviceHashes{}
	datum.Parse(parser)
	return datum
}

func (d *DeviceHashes) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		(*d)[reference] = *parser.String(reference)
	}
}

func (d *DeviceHashes) Validate(validator structure.Validator) {
	if length := len(*d); length > DeviceHashesLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, DeviceHashesLengthMaximum))
	}
	for _, deviceID := range slices.Sorted(maps.Keys(*d)) {
		validator.String(deviceID+"#", &deviceID).NotEmpty().LengthLessThanOrEqualTo(DeviceIDLengthMaximum)
		deviceHash := (*d)[deviceID]
		validator.String(deviceID, &deviceHash).NotEmpty().LengthLessThanOrEqualTo(DeviceHashLengthMaximum)
	}
}

const MetadataKeyDeviceHashes = "deviceHashes"

type DeviceHashesMetadata struct {
	DeviceHashes *DeviceHashes `json:"deviceHashes,omitempty" bson:"deviceHashes,omitempty"`
}

func (d *DeviceHashesMetadata) Parse(parser structure.ObjectParser) {
	d.DeviceHashes = ParseDeviceHashes(parser.WithReferenceObjectParser(MetadataKeyDeviceHashes))
}

func (d *DeviceHashesMetadata) Validate(validator structure.Validator) {
	if d.DeviceHashes != nil {
		d.DeviceHashes.Validate(validator.WithReference(MetadataKeyDeviceHashes))
	}
}
