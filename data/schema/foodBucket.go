package schema

import (
	"time"

	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/errors"
)

type (
	FoodBucket struct {
		Id                string    `bson:"_id,omitempty"`
		CreationTimestamp time.Time `bson:"creationTimestamp,omitempty"`
		UserId            string    `bson:"userId,omitempty" `
		Day               time.Time `bson:"day,omitempty"` // ie: 2021-09-28
		Samples           []Food    `bson:"samples"`
	}

	Food struct {
		Sample              `bson:",inline"`
		Uuid                string     `bson:"uuid,omitempty"`
		Type                string     `bson:"meal,omitempty"`
		Nutrition           Nutrition  `bson:"nutrition,omitempty"`
		Prescriptor         *string    `bson:"prescriptor,omitempty"`
		PrescribedNutrition *Nutrition `bson:"prescribedNutrition,omitempty"`
	}
	Nutrition struct {
		Carbohydrate Carb `bson:"carbohydrate,omitempty"`
	}
	Carb struct {
		Net   float64 `bson:"net,omitempty"`
		Units string  `bson:"units,omitempty"`
	}
)

func (f FoodBucket) GetId() string {
	return f.Id
}

func (f Food) GetTimestamp() time.Time {
	return f.Timestamp
}
func (f *Food) MapForFood(event *food.Food) error {
	var err error
	if event.ID != nil {
		f.Uuid = *event.ID
	}
	if event.Meal != nil {
		f.Type = *event.Meal
	}

	if event.Prescriptor != nil {
		f.Prescriptor = event.Prescriptor.Prescriptor
	}

	if event.Nutrition != nil && event.Nutrition.Carbohydrate != nil {
		f.Nutrition = Nutrition{
			Carbohydrate: Carb{
				Net:   *event.Nutrition.Carbohydrate.Net,
				Units: *event.Nutrition.Carbohydrate.Units,
			},
		}
	}

	if event.PrescribedNutrition != nil && event.PrescribedNutrition.Carbohydrate != nil {
		f.PrescribedNutrition = &Nutrition{
			Carbohydrate: Carb{
				Net:   *event.PrescribedNutrition.Carbohydrate.Net,
				Units: *event.PrescribedNutrition.Carbohydrate.Units,
			},
		}
	}

	// time infos mapping
	f.Timezone = *event.TimeZoneName
	f.TimezoneOffset = *event.TimeZoneOffset
	// what is this mess ???
	strTime := *event.Time
	f.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}
