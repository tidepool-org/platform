package factory_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/factory"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/base/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/base/basal/suspend"
	"github.com/tidepool-org/platform/data/types/base/basal/temporary"
	"github.com/tidepool-org/platform/data/types/base/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/base/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/base/blood/ketone"
	"github.com/tidepool-org/platform/data/types/base/bolus/combination"
	"github.com/tidepool-org/platform/data/types/base/bolus/extended"
	"github.com/tidepool-org/platform/data/types/base/bolus/normal"
	"github.com/tidepool-org/platform/data/types/base/calculator"
	"github.com/tidepool-org/platform/data/types/base/device/alarm"
	"github.com/tidepool-org/platform/data/types/base/device/calibration"
	"github.com/tidepool-org/platform/data/types/base/device/prime"
	"github.com/tidepool-org/platform/data/types/base/device/reservoirchange"
	"github.com/tidepool-org/platform/data/types/base/device/status"
	"github.com/tidepool-org/platform/data/types/base/device/timechange"
	"github.com/tidepool-org/platform/data/types/base/settings/pump"
	"github.com/tidepool-org/platform/data/types/base/upload"
)

type NewInvalidPropertyErrorInput struct {
	key           string
	value         string
	allowedValues []string
}

type TestInspector struct {
	GetPropertyInputs              []string
	GetPropertyOutputs             []*string
	NewMissingPropertyErrorInputs  []string
	NewMissingPropertyErrorOutputs []error
	NewInvalidPropertyErrorInputs  []NewInvalidPropertyErrorInput
	NewInvalidPropertyErrorOutputs []error
}

func (t *TestInspector) GetProperty(key string) *string {
	t.GetPropertyInputs = append(t.GetPropertyInputs, key)
	output := t.GetPropertyOutputs[0]
	t.GetPropertyOutputs = t.GetPropertyOutputs[1:]
	return output
}

func (t *TestInspector) NewMissingPropertyError(key string) error {
	t.NewMissingPropertyErrorInputs = append(t.NewMissingPropertyErrorInputs, key)
	output := t.NewMissingPropertyErrorOutputs[0]
	t.NewMissingPropertyErrorOutputs = t.NewMissingPropertyErrorOutputs[1:]
	return output
}

func (t *TestInspector) NewInvalidPropertyError(key string, value string, allowedValues []string) error {
	t.NewInvalidPropertyErrorInputs = append(t.NewInvalidPropertyErrorInputs, NewInvalidPropertyErrorInput{key, value, allowedValues})
	output := t.NewInvalidPropertyErrorOutputs[0]
	t.NewInvalidPropertyErrorOutputs = t.NewInvalidPropertyErrorOutputs[1:]
	return output
}

type TestPropertyMapInspector struct {
	PropertyMap map[string]string
}

func (t *TestPropertyMapInspector) GetProperty(key string) *string {
	value, ok := t.PropertyMap[key]
	if !ok {
		return nil
	}
	return &value
}

func (t *TestPropertyMapInspector) NewMissingPropertyError(key string) error {
	return fmt.Errorf("test: %s is missing", key)
}

func (t *TestPropertyMapInspector) NewInvalidPropertyError(key string, value string, allowedValues []string) error {
	return fmt.Errorf("test: %s is invalid", key)
}

var _ = Describe("Standard", func() {
	Context("NewNewFuncWithFunc", func() {
		It("returns nil if the datumFunc is nil", func() {
			Expect(factory.NewNewFuncWithFunc(nil)).To(BeNil())
		})

		It("returns a NewFunc if the datumFunc is not nil", func() {
			Expect(factory.NewNewFuncWithFunc(func() data.Datum { return nil })).ToNot(BeNil())
		})

		It("returns a NewFunc that returns an error if the inspector is nil", func() {
			newFunc := factory.NewNewFuncWithFunc(func() data.Datum { return nil })
			Expect(newFunc).ToNot(BeNil())
			datum, err := newFunc(nil)
			Expect(err).To(MatchError("factory: inspector is missing"))
			Expect(datum).To(BeNil())
		})

		Context("with inspector", func() {
			var testInspector *TestInspector

			BeforeEach(func() {
				testInspector = &TestInspector{}
			})

			It("returns a NewFunc that returns nil if the datumFunc returns nil", func() {
				newFunc := factory.NewNewFuncWithFunc(func() data.Datum { return nil })
				Expect(newFunc).ToNot(BeNil())
				Expect(newFunc(testInspector)).To(BeNil())
			})

			It("returns a NewFunc that returns the datum that the datumFunc returns", func() {
				testDatum := testData.NewDatum()
				newFunc := factory.NewNewFuncWithFunc(func() data.Datum { return testDatum })
				Expect(newFunc).ToNot(BeNil())
				Expect(newFunc(testInspector)).To(Equal(testDatum))
			})
		})
	})

	Context("NewNewFuncWithKeyAndMap", func() {
		var testDatum data.Datum
		var testNewFuncMap factory.NewFuncMap
		var testNewFuncAllowedValues []string
		var testInspector *TestInspector

		BeforeEach(func() {
			testDatum = testData.NewDatum()
			testNewFuncMap = factory.NewFuncMap{
				"value-datum-func-returns-datum": func(_ data.Inspector) (data.Datum, error) { return testDatum, nil },
				"value-datum-func-returns-error": func(_ data.Inspector) (data.Datum, error) { return nil, errors.New("test: datum func returns error") },
				"value-new-func-nil":             nil,
			}
			testNewFuncAllowedValues = []string{"value-datum-func-returns-datum", "value-datum-func-returns-error", "value-new-func-nil"}
			testInspector = &TestInspector{}
		})

		It("returns a NewFunc that returns the datum that the datumFunc returns", func() {
			testInspector.GetPropertyOutputs = []*string{app.StringAsPointer("value-datum-func-returns-datum")}
			newFunc := factory.NewNewFuncWithKeyAndMap("key-datum-func-returns-datum", testNewFuncMap)
			Expect(newFunc).ToNot(BeNil())
			Expect(newFunc(testInspector)).To(Equal(testDatum))
			Expect(testInspector.GetPropertyInputs).To(ConsistOf("key-datum-func-returns-datum"))
		})

		It("returns a NewFunc that returns the error that the datumFunc returns", func() {
			testInspector.GetPropertyOutputs = []*string{app.StringAsPointer("value-datum-func-returns-error")}
			newFunc := factory.NewNewFuncWithKeyAndMap("key-datum-func-returns-error", testNewFuncMap)
			Expect(newFunc).ToNot(BeNil())
			datum, err := newFunc(testInspector)
			Expect(err).To(MatchError("test: datum func returns error"))
			Expect(datum).To(BeNil())
			Expect(testInspector.GetPropertyInputs).To(ConsistOf("key-datum-func-returns-error"))
		})

		It("returns a NewFunc that returns an error if the inspector is nil", func() {
			newFunc := factory.NewNewFuncWithKeyAndMap("key-datum-func-returns-datum", testNewFuncMap)
			Expect(newFunc).ToNot(BeNil())
			datum, err := newFunc(nil)
			Expect(err).To(MatchError("factory: inspector is missing"))
			Expect(datum).To(BeNil())
			Expect(testInspector.GetPropertyInputs).To(BeEmpty())
		})

		It("returns a NewFunc that returns an error if the key is not found by the inspector", func() {
			testInspector.GetPropertyOutputs = []*string{nil}
			testInspector.NewMissingPropertyErrorOutputs = []error{errors.New("test: key not found by inspector")}
			newFunc := factory.NewNewFuncWithKeyAndMap("key-not-found-by-inspector", testNewFuncMap)
			Expect(newFunc).ToNot(BeNil())
			datum, err := newFunc(testInspector)
			Expect(err).To(MatchError("test: key not found by inspector"))
			Expect(datum).To(BeNil())
			Expect(testInspector.GetPropertyInputs).To(ConsistOf("key-not-found-by-inspector"))
			Expect(testInspector.NewMissingPropertyErrorInputs).To(ConsistOf("key-not-found-by-inspector"))
		})

		It("returns a NewFunc that returns an error if the value returned by the inspector is not found in the new func map", func() {
			testInspector.GetPropertyOutputs = []*string{app.StringAsPointer("value-new-func-not-found")}
			testInspector.NewInvalidPropertyErrorOutputs = []error{errors.New("test: value new func not found")}
			newFunc := factory.NewNewFuncWithKeyAndMap("key-new-func-not-found", testNewFuncMap)
			Expect(newFunc).ToNot(BeNil())
			datum, err := newFunc(testInspector)
			Expect(err).To(MatchError("test: value new func not found"))
			Expect(datum).To(BeNil())
			Expect(testInspector.GetPropertyInputs).To(ConsistOf("key-new-func-not-found"))
			Expect(testInspector.NewInvalidPropertyErrorInputs).To(ConsistOf(NewInvalidPropertyErrorInput{"key-new-func-not-found", "value-new-func-not-found", testNewFuncAllowedValues}))
		})

		It("returns a NewFunc that returns an error if the value returned by the inspector is nil in the new func map", func() {
			testInspector.GetPropertyOutputs = []*string{app.StringAsPointer("value-new-func-nil")}
			testInspector.NewMissingPropertyErrorOutputs = []error{errors.New("test: value new func nil")}
			newFunc := factory.NewNewFuncWithKeyAndMap("key-new-func-nil", testNewFuncMap)
			Expect(newFunc).ToNot(BeNil())
			datum, err := newFunc(testInspector)
			Expect(err).To(MatchError("test: value new func nil"))
			Expect(datum).To(BeNil())
			Expect(testInspector.GetPropertyInputs).To(ConsistOf("key-new-func-nil"))
			Expect(testInspector.NewMissingPropertyErrorInputs).To(ConsistOf("value-new-func-nil"))
		})
	})

	Context("NewStandard", func() {
		It("returns a standard without error", func() {
			Expect(factory.NewStandard()).ToNot(BeNil())
		})

		Context("with a new factory", func() {
			var standard *factory.Standard

			BeforeEach(func() {
				var err error
				standard, err = factory.NewStandard()
				Expect(err).ToNot(HaveOccurred())
				Expect(standard).ToNot(BeNil())
			})

			ValidStandardFactoryEntries := []TableEntry{
				Entry("is basal scheduled", map[string]string{"type": "basal", "deliveryType": "scheduled"}, scheduled.New()),
				Entry("is basal suspend", map[string]string{"type": "basal", "deliveryType": "suspend"}, suspend.New()),
				Entry("is basal temp", map[string]string{"type": "basal", "deliveryType": "temp"}, temporary.New()),
				Entry("is wizard", map[string]string{"type": "wizard"}, calculator.New()),
				Entry("is bolus dual/square", map[string]string{"type": "bolus", "subType": "dual/square"}, combination.New()),
				Entry("is bolus square", map[string]string{"type": "bolus", "subType": "square"}, extended.New()),
				Entry("is bolus normal", map[string]string{"type": "bolus", "subType": "normal"}, normal.New()),
				Entry("is cbg", map[string]string{"type": "cbg"}, continuous.New()),
				Entry("is deviceEvent alarm", map[string]string{"type": "deviceEvent", "subType": "alarm"}, alarm.New()),
				Entry("is deviceEvent calibration", map[string]string{"type": "deviceEvent", "subType": "calibration"}, calibration.New()),
				Entry("is deviceEvent prime", map[string]string{"type": "deviceEvent", "subType": "prime"}, prime.New()),
				Entry("is deviceEvent reservoirChange", map[string]string{"type": "deviceEvent", "subType": "reservoirChange"}, reservoirchange.New()),
				Entry("is deviceEvent status", map[string]string{"type": "deviceEvent", "subType": "status"}, status.New()),
				Entry("is deviceEvent timeChange", map[string]string{"type": "deviceEvent", "subType": "timeChange"}, timechange.New()),
				Entry("is bloodKetone", map[string]string{"type": "bloodKetone"}, ketone.New()),
				Entry("is pumpSettings", map[string]string{"type": "pumpSettings"}, pump.New()),
				Entry("is smbg", map[string]string{"type": "smbg"}, selfmonitored.New()),
				Entry("is upload", map[string]string{"type": "upload"}, upload.New()),
			}

			InvalidStandardFactoryEntries := []TableEntry{
				Entry("is basal unknown", map[string]string{"type": "basal", "deliveryType": "unknown"}, "test: deliveryType is invalid"),
				Entry("is bolus unknown", map[string]string{"type": "bolus", "subType": "unknown"}, "test: subType is invalid"),
				Entry("is deviceEvent unknown", map[string]string{"type": "deviceEvent", "subType": "unknown"}, "test: subType is invalid"),
				Entry("is unknown", map[string]string{"type": "unknown"}, "test: type is invalid"),
			}

			Context("New", func() {
				DescribeTable("returns expected datum and no error when",
					func(propertyMap map[string]string, assignable interface{}) {
						datum, err := standard.New(&TestPropertyMapInspector{propertyMap})
						Expect(err).ToNot(HaveOccurred())
						Expect(datum).ToNot(BeNil())
						Expect(datum).To(BeAssignableToTypeOf(assignable))
					},
					ValidStandardFactoryEntries...,
				)

				DescribeTable("returns no datum and an error when",
					func(propertyMap map[string]string, matchError string) {
						datum, err := standard.New(&TestPropertyMapInspector{propertyMap})
						Expect(err).To(MatchError(matchError))
						Expect(datum).To(BeNil())
					},
					InvalidStandardFactoryEntries...,
				)
			})

			Context("Init", func() {
				DescribeTable("returns expected datum and no error when",
					func(propertyMap map[string]string, assignable interface{}) {
						datum, err := standard.Init(&TestPropertyMapInspector{propertyMap})
						Expect(err).ToNot(HaveOccurred())
						Expect(datum).ToNot(BeNil())
						Expect(datum).To(BeAssignableToTypeOf(assignable))
					},
					ValidStandardFactoryEntries...,
				)

				DescribeTable("returns no datum and an error when",
					func(propertyMap map[string]string, matchError string) {
						datum, err := standard.Init(&TestPropertyMapInspector{propertyMap})
						Expect(err).To(MatchError(matchError))
						Expect(datum).To(BeNil())
					},
					InvalidStandardFactoryEntries...,
				)
			})
		})
	})
})
