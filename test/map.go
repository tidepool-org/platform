package test

import (
	"github.com/onsi/gomega"
	gomegaGstruct "github.com/onsi/gomega/gstruct"
	gomegaTypes "github.com/onsi/gomega/types"
)

func MatchMap[K comparable, V any](expected map[K]V) gomegaTypes.GomegaMatcher {
	keys := gomegaGstruct.Keys{}
	for k, v := range expected {
		keys[k] = gomega.Equal(v)
	}
	return gomegaGstruct.MatchAllKeys(keys)
}
