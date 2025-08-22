package test

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"

	"time"

	"github.com/tidepool-org/platform/consent"
)

const (
	ConsentType        = "test_consent"
	AnotherConsentType = "another_test_consent"
)

var ConsentV1 = &consent.Consent{
	Type:        ConsentType,
	Version:     1,
	Content:     "# Test Consent - version 1",
	ContentType: "markdown",
	CreatedTime: time.Now().UTC(),
}

var ConsentV2 = &consent.Consent{
	Type:        ConsentType,
	Version:     2,
	Content:     "# Test Consent - version 2",
	ContentType: "markdown",
	CreatedTime: time.Now().UTC(),
}

var AnotherConsentV1 = &consent.Consent{
	Type:        AnotherConsentType,
	Version:     1,
	Content:     "# Another Test Consent - version 1",
	ContentType: "markdown",
	CreatedTime: time.Now().UTC(),
}

var MockBDDPConsentV1 = &consent.Consent{
	Type:        "big_data_donation_project",
	Version:     1,
	Content:     "# Tidepool BDDP Consent - version 1",
	ContentType: "markdown",
	CreatedTime: time.Now().UTC(),
}

func MatchConsent(cons consent.Consent) types.GomegaMatcher {
	return MatchFields(IgnoreExtras, Fields{
		"Type":        Equal(cons.Type),
		"Version":     Equal(cons.Version),
		"Content":     Equal(cons.Content),
		"ContentType": Equal(cons.ContentType),
		"CreatedTime": BeTemporally("~", cons.CreatedTime, time.Second),
	})
}
