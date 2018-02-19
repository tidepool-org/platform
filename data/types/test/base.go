package test

import (
	"math/rand"

	"github.com/tidepool-org/platform/test"
)

const (
	AnnotationsMaximum = 3
	PropertiesMaximum  = 3
)

func NewAnnotations() *[]map[string]interface{} {
	annotations := make([]map[string]interface{}, rand.Intn(AnnotationsMaximum)+1)
	for index := range annotations {
		annotations[index] = *NewPropertyMap()
	}
	return &annotations
}

func NewClockDriftOffset() int {
	return -14400000 + rand.Intn(14400000+3600000)
}

func NewConversionOffset() int {
	return -9999999999 + rand.Intn(9999999999+9999999999)
}

func NewPayload() *map[string]interface{} {
	return NewPropertyMap()
}

func NewTimezoneOffset() int {
	return -4440 + rand.Intn(4440+6960)
}

func NewVersion() int {
	return rand.Intn(10)
}

func NewPropertyMap() *map[string]interface{} {
	propertyMap := map[string]interface{}{}
	for index := rand.Intn(PropertiesMaximum); index >= 0; index-- {
		propertyMap[test.NewVariableString(1, 8, test.CharsetAlpha)] = test.NewVariableString(0, 16, test.CharsetAlpha)
	}
	return &propertyMap
}
