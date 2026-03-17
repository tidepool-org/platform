package work

import "fmt"

const Domain = "org.tidepool.oura.data"

func SerialIDFromProviderSessionID(providerSessionID string) string {
	return fmt.Sprintf("%s:%s", Domain, providerSessionID)
}
