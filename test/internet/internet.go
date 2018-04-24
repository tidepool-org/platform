package internet

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/tidepool-org/platform/test"
)

const SubDomainsMaximum = 3

var tlds = []string{
	"com",
	"edu",
	"gov",
	"info",
	"net",
	"org",
}

func NewTLD() string {
	return tlds[rand.Intn(len(tlds))]
}

func NewDomain() string {
	return fmt.Sprintf("%s.%s", NewSubDomains(), NewTLD())
}

func NewReverseDomain() string {
	return fmt.Sprintf("%s.%s", NewTLD(), NewSubDomains())
}

func NewSubDomains() string {
	subDomains := make([]string, rand.Intn(SubDomainsMaximum)+1)
	for index := range subDomains {
		subDomains[index] = test.NewVariableString(1, 8, test.CharsetAlpha)
	}
	return strings.Join(subDomains, ".")
}

func NewEmail() string {
	return fmt.Sprintf("%s+%s@%s", test.NewVariableString(1, 16, test.CharsetAlpha), test.NewVariableString(1, 8, test.CharsetNumeric), NewDomain())
}

func NewSemanticVersion() string {
	return fmt.Sprintf("%d.%d.%d", rand.Intn(10), rand.Intn(10), rand.Intn(10))
}
