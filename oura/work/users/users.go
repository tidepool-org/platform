package users

import (
	"fmt"
)

const Domain = "org.tidepool.oura.work.users"

func GroupIDFromProviderSessionID(providerSessionID string) string {
	return fmt.Sprintf("%s:%s", Domain, providerSessionID)
}

func SerialIDFromProviderSessionID(providerSessionID string) string {
	return fmt.Sprintf("%s:%s", Domain, providerSessionID)
}
