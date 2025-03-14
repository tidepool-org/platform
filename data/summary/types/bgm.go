package types

import (
	"encoding/json"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"go.mongodb.org/mongo-driver/bson"
)

type BGMPeriods struct {
	GlucosePeriods
}

func (*BGMPeriods) GetType() string {
	return SummaryTypeBGM
}

func (*BGMPeriods) GetDeviceDataTypes() []string {
	return []string{selfmonitored.Type}
}

func (p *BGMPeriods) MarshalJSON() ([]byte, error) {
	if p == nil {
		return bson.Marshal(GlucosePeriods{})
	}
	return json.Marshal(p.GlucosePeriods)
}

func (p *BGMPeriods) MarshalBSON() ([]byte, error) {
	if p == nil {
		return bson.Marshal(GlucosePeriods{})
	}
	return bson.Marshal(p.GlucosePeriods)
}

func (p *BGMPeriods) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.GlucosePeriods)
}

func (p *BGMPeriods) UnmarshalBSON(data []byte) error {
	return bson.Unmarshal(data, &p.GlucosePeriods)

}
