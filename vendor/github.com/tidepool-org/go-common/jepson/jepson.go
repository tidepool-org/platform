package jepson

import (
	"encoding/json"
	"time"
)

func JSONString(data []byte) (string, error) {
	var retVal string
	err := json.Unmarshal(data, &retVal)
	return retVal, err
}

type Duration time.Duration

func (jd *Duration) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err != nil {
		return err
	}

	dur, err := time.ParseDuration(asString)
	if err != nil {
		return err
	}

	*jd = Duration(dur)
	return nil
}

func (jd *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(*jd).String())
}
