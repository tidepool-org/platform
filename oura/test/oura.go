package test

import (
	"math/rand/v2"
	"slices"

	metadataTest "github.com/tidepool-org/platform/metadata/test"
	netTest "github.com/tidepool-org/platform/net/test"
	oura "github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

func RandomCallbackURL() string {
	return testHttp.NewURL().String()
}

func RandomVerificationToken() string {
	return test.RandomString()
}

func RandomChallenge() string {
	return test.RandomString()
}

func RandomDataType() string {
	return test.RandomStringFromArray(oura.DataTypes())
}

func RandomEventType() string {
	return test.RandomStringFromArray(oura.EventTypes())
}

func RandomID() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomUserID() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomNextToken() string {
	return test.RandomStringFromCharset(test.CharsetAlphaNumeric)
}

func RandomScope() []string {
	return test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, 3, oura.Scopes())
}

func RandomCreateSubscription(options ...test.Option) *oura.CreateSubscription {
	return &oura.CreateSubscription{
		CallbackURL:       pointer.From(RandomCallbackURL()),
		VerificationToken: pointer.From(RandomVerificationToken()),
		DataType:          pointer.From(RandomDataType()),
		EventType:         pointer.From(RandomEventType()),
	}
}

func CloneCreateSubscription(datum *oura.CreateSubscription) *oura.CreateSubscription {
	if datum == nil {
		return nil
	}
	return &oura.CreateSubscription{
		CallbackURL:       pointer.Clone(datum.CallbackURL),
		VerificationToken: pointer.Clone(datum.VerificationToken),
		DataType:          pointer.Clone(datum.DataType),
		EventType:         pointer.Clone(datum.EventType),
	}
}

func NewObjectFromCreateSubscription(datum *oura.CreateSubscription, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.CallbackURL != nil {
		object["callback_url"] = test.NewObjectFromString(*datum.CallbackURL, format)
	}
	if datum.VerificationToken != nil {
		object["verification_token"] = test.NewObjectFromString(*datum.VerificationToken, format)
	}
	if datum.DataType != nil {
		object["data_type"] = test.NewObjectFromString(*datum.DataType, format)
	}
	if datum.EventType != nil {
		object["event_type"] = test.NewObjectFromString(*datum.EventType, format)
	}
	return object
}

func RandomUpdateSubscription(options ...test.Option) *oura.UpdateSubscription {
	return &oura.UpdateSubscription{
		CallbackURL:       pointer.From(RandomCallbackURL()),
		VerificationToken: pointer.From(RandomVerificationToken()),
		DataType:          pointer.From(RandomDataType()),
		EventType:         pointer.From(RandomEventType()),
	}
}

func CloneUpdateSubscription(datum *oura.UpdateSubscription) *oura.UpdateSubscription {
	if datum == nil {
		return nil
	}
	return &oura.UpdateSubscription{
		CallbackURL:       pointer.Clone(datum.CallbackURL),
		VerificationToken: pointer.Clone(datum.VerificationToken),
		DataType:          pointer.Clone(datum.DataType),
		EventType:         pointer.Clone(datum.EventType),
	}
}

func NewObjectFromUpdateSubscription(datum *oura.UpdateSubscription, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.CallbackURL != nil {
		object["callback_url"] = test.NewObjectFromString(*datum.CallbackURL, format)
	}
	if datum.VerificationToken != nil {
		object["verification_token"] = test.NewObjectFromString(*datum.VerificationToken, format)
	}
	if datum.DataType != nil {
		object["data_type"] = test.NewObjectFromString(*datum.DataType, format)
	}
	if datum.EventType != nil {
		object["event_type"] = test.NewObjectFromString(*datum.EventType, format)
	}
	return object
}

func RandomSubscription(options ...test.Option) *oura.Subscription {
	return &oura.Subscription{
		ID:             pointer.From(RandomID()),
		CallbackURL:    pointer.From(RandomCallbackURL()),
		DataType:       pointer.From(RandomDataType()),
		EventType:      pointer.From(RandomEventType()),
		ExpirationTime: pointer.From(test.RandomTimeAfterNow().UTC().Format(oura.SubscriptionExpirationTimeFormat)),
	}
}

func CloneSubscription(datum *oura.Subscription) *oura.Subscription {
	if datum == nil {
		return nil
	}
	return &oura.Subscription{
		ID:             pointer.Clone(datum.ID),
		CallbackURL:    pointer.Clone(datum.CallbackURL),
		DataType:       pointer.Clone(datum.DataType),
		EventType:      pointer.Clone(datum.EventType),
		ExpirationTime: pointer.Clone(datum.ExpirationTime),
	}
}

func NewObjectFromSubscription(datum *oura.Subscription, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, format)
	}
	if datum.CallbackURL != nil {
		object["callback_url"] = test.NewObjectFromString(*datum.CallbackURL, format)
	}
	if datum.DataType != nil {
		object["data_type"] = test.NewObjectFromString(*datum.DataType, format)
	}
	if datum.EventType != nil {
		object["event_type"] = test.NewObjectFromString(*datum.EventType, format)
	}
	if datum.ExpirationTime != nil {
		object["expiration_time"] = test.NewObjectFromString(*datum.ExpirationTime, format)
	}
	return object
}

// RandomSubscriptions should ensure unique combinations of data types and event types
func RandomSubscriptions(options ...test.Option) oura.Subscriptions {
	dataTypes := oura.DataTypes()
	eventTypes := oura.EventTypes()
	dataTypesCount := len(dataTypes)
	eventTypesCount := len(eventTypes)
	subscriptionsCount := dataTypesCount * eventTypesCount
	offsets := rand.Perm(subscriptionsCount)
	subscriptions := make(oura.Subscriptions, test.RandomIntFromRange(1, subscriptionsCount))
	for index := range subscriptions {
		subscription := RandomSubscription(options...)
		subscription.DataType = pointer.From(dataTypes[offsets[index]%dataTypesCount])
		subscription.EventType = pointer.From(eventTypes[offsets[index]/dataTypesCount])
		subscriptions[index] = subscription
	}
	return subscriptions
}

func CloneSubscriptions(datum *oura.Subscriptions) oura.Subscriptions {
	if datum == nil {
		return nil
	}
	cloned := make(oura.Subscriptions, len(*datum))
	for index, datum := range *datum {
		cloned[index] = CloneSubscription(datum)
	}
	return cloned
}

func NewArrayFromSubscriptions(datum *oura.Subscriptions, format test.ObjectFormat) []any {
	if datum == nil {
		return nil
	}
	array := make([]any, len(*datum))
	for index, datum := range *datum {
		array[index] = NewObjectFromSubscription(datum, format)
	}
	return array
}

func RandomPersonalInfo(options ...test.Option) *oura.PersonalInfo {
	return &oura.PersonalInfo{
		ID:            pointer.From(RandomUserID()),
		Age:           test.RandomOptional(test.RandomInt, options...),
		Weight:        test.RandomOptional(test.RandomFloat64, options...),
		Height:        test.RandomOptional(test.RandomFloat64, options...),
		BiologicalSex: test.RandomOptional(test.RandomString, options...),
		Email:         test.RandomOptional(netTest.RandomEmail, options...),
	}
}

func ClonePersonalInfo(datum *oura.PersonalInfo) *oura.PersonalInfo {
	if datum == nil {
		return nil
	}
	return &oura.PersonalInfo{
		ID:            pointer.Clone(datum.ID),
		Age:           pointer.Clone(datum.Age),
		Weight:        pointer.Clone(datum.Weight),
		Height:        pointer.Clone(datum.Height),
		BiologicalSex: pointer.Clone(datum.BiologicalSex),
		Email:         pointer.Clone(datum.Email),
	}
}

func NewObjectFromPersonalInfo(datum *oura.PersonalInfo, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, format)
	}
	if datum.Age != nil {
		object["age"] = test.NewObjectFromInt(*datum.Age, format)
	}
	if datum.Weight != nil {
		object["weight"] = test.NewObjectFromFloat64(*datum.Weight, format)
	}
	if datum.Height != nil {
		object["height"] = test.NewObjectFromFloat64(*datum.Height, format)
	}
	if datum.BiologicalSex != nil {
		object["biological_sex"] = test.NewObjectFromString(*datum.BiologicalSex, format)
	}
	if datum.Email != nil {
		object["email"] = test.NewObjectFromString(*datum.Email, format)
	}
	return object
}

func RandomPagination(options ...test.Option) *oura.Pagination {
	return &oura.Pagination{
		NextToken: test.RandomOptional(RandomNextToken, options...),
	}
}

func ClonePagination(datum *oura.Pagination) *oura.Pagination {
	if datum == nil {
		return nil
	}
	return &oura.Pagination{
		NextToken: pointer.Clone(datum.NextToken),
	}
}

func NewObjectFromPagination(datum *oura.Pagination, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.NextToken != nil {
		object["next_token"] = test.NewObjectFromString(*datum.NextToken, format)
	}
	return object
}

func RandomDataResponse(options ...test.Option) *oura.DataResponse {
	data := make([]any, test.RandomIntFromRange(1, 3))
	for index := range data {
		data[index] = RandomDatum(options...)
	}
	return &oura.DataResponse{
		Data:       data,
		Pagination: *RandomPagination(options...),
	}
}

func CloneDataResponse(datum *oura.DataResponse) *oura.DataResponse {
	if datum == nil {
		return nil
	}
	return &oura.DataResponse{
		Data:       slices.Clone(datum.Data),
		Pagination: *ClonePagination(&datum.Pagination),
	}
}

func NewObjectFromDataResponse(datum *oura.DataResponse, format test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := NewObjectFromPagination(&datum.Pagination, format)
	if datum.Data != nil {
		object["data"] = slices.Clone(datum.Data)
	}
	return object
}

func RandomDatum(options ...test.Option) oura.Datum {
	return metadataTest.RandomMetadataMap()
}

func CloneDatum(datum oura.Datum) oura.Datum {
	return metadataTest.CloneMetadataMap(datum)
}

func NewObjectFromDatum(datum oura.Datum, format test.ObjectFormat) map[string]any {
	return metadataTest.NewObjectFromMetadataMap(datum, format)
}

func RandomData(options ...test.Option) []any {
	datum := make([]any, test.RandomIntFromRange(1, 3))
	for index := range datum {
		datum[index] = RandomDatum(options...)
	}
	return datum
}
