package internet

import (
	"fmt"
	"math/rand"

	"github.com/tidepool-org/platform/test"
)

var TLDs = []string{
	"com",
	"edu",
	"gov",
	"info",
	"net",
	"org",
}

func NewTLD() string {
	return TLDs[rand.Intn(len(TLDs))]
}

func NewDomain() string {
	return fmt.Sprintf("%s.%s", test.NewVariableString(1, 8, test.CharsetAlpha), NewTLD())
}

func NewEmail() string {
	return fmt.Sprintf("%s+%s@%s", test.NewVariableString(1, 16, test.CharsetAlpha), test.NewVariableString(1, 8, test.CharsetNumeric), NewDomain())
}
