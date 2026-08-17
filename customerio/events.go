package customerio

const (
	DataSourceStateChangedEventType = "data_source_state_changed"
)

type DataSourceStateChangedEvent struct {
	ProviderName string `json:"provider_name" bson:"provider_name"`
	State        string `json:"state" bson:"state"`
}
