package types

import (
	"encoding/json"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"go.mongodb.org/mongo-driver/bson"
)

type CGMPeriods struct {
	GlucosePeriods `json:",inline" bson:",inline"`
}

func (*CGMPeriods) GetType() string {
	return SummaryTypeCGM
}

func (*CGMPeriods) GetDeviceDataTypes() []string {
	return []string{continuous.Type}
}

func (p *CGMPeriods) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GlucosePeriods)
}

func (p *CGMPeriods) MarshalBSON() ([]byte, error) {
	return json.Marshal(p.GlucosePeriods)
}

func (p *CGMPeriods) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.GlucosePeriods)
}

func (p *CGMPeriods) UnmarshalBSON(data []byte) error {
	return bson.Unmarshal(data, &p.GlucosePeriods)

}
