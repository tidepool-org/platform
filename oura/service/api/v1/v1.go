package v1

import (
	dataService "github.com/tidepool-org/platform/data/service"
)

func Routes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/partners/oura/event", Subscription),
		dataService.Post("/v1/partners/oura/event", Event),
	}
}
