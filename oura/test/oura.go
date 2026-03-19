package test

import (
	netTest "github.com/tidepool-org/platform/net/test"
	oura "github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCallbackURL() string {
	return test.RandomString()
}

func RandomVerificationToken() string {
	return test.RandomString()
}

func RandomDataType() string {
	return test.RandomStringFromArray(oura.DataTypes())
}

func RandomEventType() string {
	return test.RandomStringFromArray(oura.EventTypes())
}

func RandomID() string {
	return test.RandomString()
}

func RandomUserID() string {
	return test.RandomString()
}

func RandomCreateSubscription(options ...test.Option) *oura.CreateSubscription {
	return &oura.CreateSubscription{
		CallbackURL:       pointer.FromString(RandomCallbackURL()),
		VerificationToken: pointer.FromString(RandomVerificationToken()),
		DataType:          pointer.FromString(RandomDataType()),
		EventType:         pointer.FromString(RandomEventType()),
	}
}

func CloneCreateSubscription(datum *oura.CreateSubscription) *oura.CreateSubscription {
	if datum == nil {
		return nil
	}
	return &oura.CreateSubscription{
		CallbackURL:       pointer.CloneString(datum.CallbackURL),
		VerificationToken: pointer.CloneString(datum.VerificationToken),
		DataType:          pointer.CloneString(datum.DataType),
		EventType:         pointer.CloneString(datum.EventType),
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
		CallbackURL:       pointer.FromString(RandomCallbackURL()),
		VerificationToken: pointer.FromString(RandomVerificationToken()),
		DataType:          pointer.FromString(RandomDataType()),
		EventType:         pointer.FromString(RandomEventType()),
	}
}

func CloneUpdateSubscription(datum *oura.UpdateSubscription) *oura.UpdateSubscription {
	if datum == nil {
		return nil
	}
	return &oura.UpdateSubscription{
		CallbackURL:       pointer.CloneString(datum.CallbackURL),
		VerificationToken: pointer.CloneString(datum.VerificationToken),
		DataType:          pointer.CloneString(datum.DataType),
		EventType:         pointer.CloneString(datum.EventType),
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
		ID:             pointer.FromString(RandomID()),
		CallbackURL:    pointer.FromString(RandomCallbackURL()),
		DataType:       pointer.FromString(RandomDataType()),
		EventType:      pointer.FromString(RandomEventType()),
		ExpirationTime: pointer.FromTime(test.RandomTimeAfterNow().UTC()),
	}
}

func CloneSubscription(datum *oura.Subscription) *oura.Subscription {
	if datum == nil {
		return nil
	}
	return &oura.Subscription{
		ID:             pointer.CloneString(datum.ID),
		CallbackURL:    pointer.CloneString(datum.CallbackURL),
		DataType:       pointer.CloneString(datum.DataType),
		EventType:      pointer.CloneString(datum.EventType),
		ExpirationTime: pointer.CloneTime(datum.ExpirationTime),
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
		object["expiration_time"] = test.NewObjectFromTime(*datum.ExpirationTime, format)
	}
	return object
}

func RandomSubscriptions(options ...test.Option) oura.Subscriptions {
	subscriptions := make(oura.Subscriptions, test.RandomIntFromRange(1, 3))
	for index := range subscriptions {
		subscriptions[index] = RandomSubscription(options...)
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
		ID:            pointer.FromString(RandomUserID()),
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
		ID:            pointer.CloneString(datum.ID),
		Age:           pointer.CloneInt(datum.Age),
		Weight:        pointer.CloneFloat64(datum.Weight),
		Height:        pointer.CloneFloat64(datum.Height),
		BiologicalSex: pointer.CloneString(datum.BiologicalSex),
		Email:         pointer.CloneString(datum.Email),
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
