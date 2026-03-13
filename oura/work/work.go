package work

import "fmt"

const Domain = "org.tidepool.oura"

func GroupIDFromProviderSessionID(providerSessionID string) string {
	return fmt.Sprintf("%s:%s", Domain, providerSessionID)
}
