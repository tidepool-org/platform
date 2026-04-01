package data

import (
	"fmt"
)

const Domain = "org.tidepool.oura.work.data"

func GroupIDFromDataSourceID(dataSrcID string) string {
	return fmt.Sprintf("%s:%s", Domain, dataSrcID)
}

func SerialIDFromDataSourceID(dataSrcID string) string {
	return fmt.Sprintf("%s:%s", Domain, dataSrcID)
}
