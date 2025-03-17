package types

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
)

type CGMPeriods struct {
	GlucosePeriods
}

func (*CGMPeriods) GetType() string {
	return SummaryTypeCGM
}

func (*CGMPeriods) GetDeviceDataTypes() []string {
	return []string{continuous.Type}
}

func (p *CGMPeriods) MarshalJSON() ([]byte, error) {
	if p == nil {
		return bson.Marshal(GlucosePeriods{})
	}
	return json.Marshal(p.GlucosePeriods)
}

func (p *CGMPeriods) MarshalBSON() ([]byte, error) {
	if p == nil {
		return bson.Marshal(GlucosePeriods{})
	}
	return bson.Marshal(p.GlucosePeriods)
}

func (p *CGMPeriods) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &p.GlucosePeriods)
}

func (p *CGMPeriods) UnmarshalBSON(data []byte) error {
	return bson.Unmarshal(data, &p.GlucosePeriods)
}
