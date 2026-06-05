package webhook

import (
	"fmt"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/crypto"
)

const EventPath = "/event"

func CallbackURLForEvent(partnerURL string, eventType string, dataType string) string {
	return client.ConstructURL(partnerURL, EventPath, eventType, dataType)
}

func VerificationTokenForCallbackURL(callbackURL string, partnerSecret string) string {
	return crypto.HexEncodedSHA256Hash(fmt.Appendf(nil, "%s:%s", callbackURL, partnerSecret))
}
