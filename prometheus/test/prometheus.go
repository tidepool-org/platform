package test

import (
	"github.com/prometheus/client_golang/prometheus"
	prometheusModel "github.com/prometheus/client_model/go"

	"github.com/tidepool-org/platform/test"
)

func RandomMetricName() string {
	return test.RandomStringFromRangeAndCharset(8, 16, test.CharsetLowercase)
}

func RandomMetricHelp() string {
	return test.RandomString()
}

func MetricFamilyFromName(name string) *prometheusModel.MetricFamily {
	metricFamilies := test.Must(prometheus.DefaultGatherer.Gather())
	for _, metricFamily := range metricFamilies {
		if metricFamily.GetName() == name {
			return metricFamily
		}
	}
	return nil
}

func LabelPairsToMap(labelPairs []*prometheusModel.LabelPair) map[string]string {
	labels := map[string]string{}
	for _, labelPair := range labelPairs {
		labels[labelPair.GetName()] = labelPair.GetValue()
	}
	return labels
}
