package udi

import (
	"fmt"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/cluster"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

// Approved identification of the Tidepool Data Platform; the device identifier keeps its leading zeros for the FDA
const (
	DeviceIdentifier    = "00884400418012"
	ProductName         = "Tidepool Data Platform"
	ManufacturerName    = "Tidepool Project"
	ManufacturerAddress = "555 Bryant St., #429, Palo Alto, CA 94301"

	ReleaseIdentifierLength      = 7
	SoftwareVersionLengthMaximum = 20
	ClientVersionLengthMaximum   = SoftwareVersionLengthMaximum - len("-") - ReleaseIdentifierLength
)

// GS1 parentheses delimit application identifiers, so the client version is limited to version characters
var ClientVersionExpression = regexp.MustCompile(`^[0-9A-Za-z.+-]+$`)

type Request struct {
	ClientVersion *string `json:"clientVersion,omitempty"`
}

func NewRequest() *Request {
	return &Request{}
}

func (r *Request) Parse(parser structure.ObjectParser) {
	r.ClientVersion = parser.String("clientVersion")
}

func (r *Request) Validate(validator structure.Validator) {
	validator.String("clientVersion", r.ClientVersion).Exists().LengthInRange(1, ClientVersionLengthMaximum).Matches(ClientVersionExpression)
}

type UDI struct {
	ProductName         string `json:"productName"`
	ManufacturerName    string `json:"manufacturerName"`
	ManufacturerAddress string `json:"manufacturerAddress"`
	ReleaseDate         string `json:"releaseDate"`
	ReleaseIdentifier   string `json:"releaseIdentifier"`
	SoftwareVersion     string `json:"softwareVersion"`
	DeviceIdentifier    string `json:"deviceIdentifier"`
	UDI                 string `json:"udi"`
}

// New formats the UDI with GS1 application identifiers (01) device identifier, (11) release date as YYMMDD
// in UTC, and (8012) software version as the client version and the release identifier
func New(release *cluster.Release, clientVersion string) (*UDI, error) {
	if release == nil {
		return nil, errors.New("release is missing")
	}
	if release.Time.IsZero() {
		return nil, errors.New("release time is missing")
	}
	if len(release.Revision) < ReleaseIdentifierLength {
		return nil, errors.Newf("release revision %q is invalid", release.Revision)
	}
	if clientVersion == "" {
		return nil, errors.New("client version is missing")
	}
	if len(clientVersion) > ClientVersionLengthMaximum || !ClientVersionExpression.MatchString(clientVersion) {
		return nil, errors.Newf("client version %q is invalid", clientVersion)
	}

	releaseIdentifier := release.Revision[:ReleaseIdentifierLength]
	softwareVersion := clientVersion + "-" + releaseIdentifier

	releaseTime := release.Time.UTC()
	return &UDI{
		ProductName:         ProductName,
		ManufacturerName:    ManufacturerName,
		ManufacturerAddress: ManufacturerAddress,
		ReleaseDate:         releaseTime.Format(time.DateOnly),
		ReleaseIdentifier:   releaseIdentifier,
		SoftwareVersion:     softwareVersion,
		DeviceIdentifier:    DeviceIdentifier,
		UDI:                 fmt.Sprintf("(01)%s(11)%s(8012)%s", DeviceIdentifier, releaseTime.Format("060102"), softwareVersion),
	}, nil
}
