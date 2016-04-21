package main

import (
	"log"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/dataservices"
)

func main() {
	port, err := config.FromEnv("TIDEPOOL_DATASERVICES_PORT")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(dataservices.NewDataServiceClient().Run(port))
}
