package types

import (
	"encoding/json"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"go.mongodb.org/mongo-driver/bson"
)

type BGMPeriods struct {
	GlucosePeriods `json:",inline" bson:",inline"`
}

func (*BGMPeriods) GetType() string {
	return SummaryTypeBGM
}

func (*BGMPeriods) GetDeviceDataTypes() []string {
	return []string{selfmonitored.Type}
}

func (p *BGMPeriods) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GlucosePeriods)
}

func (p *BGMPeriods) MarshalBSON() ([]byte, error) {
	return json.Marshal(p.GlucosePeriods)
}

func (p *BGMPeriods) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.GlucosePeriods)
}

func (p *BGMPeriods) UnmarshalBSON(data []byte) error {
	return bson.Unmarshal(data, &p.GlucosePeriods)

}
