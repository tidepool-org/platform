package v1

import (
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/oura"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
)

func Routes() []dataService.Route {
	webhookPath := oura.PartnerPathPrefix + ouraWebhook.WebhookPathEvent
	return []dataService.Route{
		dataService.Get(webhookPath, Subscription),
		dataService.Post(webhookPath, Event),
	}
}
