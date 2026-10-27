package revision_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/revision"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	dataTypesActivityPhysical "github.com/tidepool-org/platform/data/types/activity/physical"
	dataTypesAlert "github.com/tidepool-org/platform/data/types/alert"
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	dataTypesBasalSuspend "github.com/tidepool-org/platform/data/types/basal/suspend"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	dataTypesBloodGlucoseContinuous "github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseSelfmonitored "github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	dataTypesBloodKetone "github.com/tidepool-org/platform/data/types/blood/ketone"
	dataTypesBolusAutomated "github.com/tidepool-org/platform/data/types/bolus/automated"
	dataTypesBolusCombination "github.com/tidepool-org/platform/data/types/bolus/combination"
	dataTypesBolusExtended "github.com/tidepool-org/platform/data/types/bolus/extended"
	dataTypesBolusNormal "github.com/tidepool-org/platform/data/types/bolus/normal"
	dataTypesCalculator "github.com/tidepool-org/platform/data/types/calculator"
	dataTypesDeviceAlarm "github.com/tidepool-org/platform/data/types/device/alarm"
	dataTypesDeviceCalibration "github.com/tidepool-org/platform/data/types/device/calibration"
	dataTypesDeviceOverrideSettingsPump "github.com/tidepool-org/platform/data/types/device/override/settings/pump"
	dataTypesDevicePrime "github.com/tidepool-org/platform/data/types/device/prime"
	dataTypesDeviceReservoirchange "github.com/tidepool-org/platform/data/types/device/reservoirchange"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	dataTypesDosingDecision "github.com/tidepool-org/platform/data/types/dosingdecision"
	dataTypesFood "github.com/tidepool-org/platform/data/types/food"
	dataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin"
	dataTypesSettingsCGM "github.com/tidepool-org/platform/data/types/settings/cgm"
	dataTypesSettingsController "github.com/tidepool-org/platform/data/types/settings/controller"
	dataTypesSettingsPump "github.com/tidepool-org/platform/data/types/settings/pump"
	dataTypesStateReported "github.com/tidepool-org/platform/data/types/state/reported"
	dataTypesStatusController "github.com/tidepool-org/platform/data/types/status/controller"
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	dataTypesWater "github.com/tidepool-org/platform/data/types/water"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func fromJSON(text string) any {
	var value any
	Expect(json.Unmarshal([]byte(text), &value)).To(Succeed())
	return value
}

func mapFromJSON(text string) map[string]any {
	return fromJSON(text).(map[string]any)
}

func randomTime() time.Time {
	return test.RandomTime().UTC()
}

// storedScheduled is a scheduled basal as its stored document decodes, canonical and projected, next to the datum.
func storedScheduled() (*dataTypesBasalScheduled.Scheduled, map[string]any) {
	datum := dataTypesBasalScheduled.New()
	datum.Base = *dataTypesTest.RandomBase()
	datum.Type = "basal"
	datum.DeliveryType = "scheduled"
	datum.Duration = pointer.FromInt(3600000)
	datum.DurationExpected = pointer.FromInt(7200000)
	datum.Rate = pointer.FromFloat64(0.85)
	datum.ScheduleName = pointer.FromString("Weekday")
	document, err := revision.ToMap(datum)
	Expect(err).ToNot(HaveOccurred())
	return datum, revision.Canonicalize(revision.Project(document))
}

var _ = Describe("Revision", func() {
	Context("ToMap", func() {
		It("returns an error when the datum is missing", func() {
			document, err := revision.ToMap(nil)
			Expect(err).To(MatchError("datum is missing"))
			Expect(document).To(BeNil())
		})

		It("returns the stored shape of a datum with times, nested documents and arrays as plain values", func() {
			datum := dataTypesBasalScheduled.New()
			datum.Base = *dataTypesTest.RandomBase()
			datum.Rate = pointer.FromFloat64(1.25)
			datum.Revisions = dataTest.RandomRevisions()
			datum.VersionInternal = 3
			document, err := revision.ToMap(datum)
			Expect(err).ToNot(HaveOccurred())
			Expect(document["time"]).To(BeAssignableToTypeOf(time.Time{}))
			Expect(document["time"].(time.Time)).To(BeTemporally("==", *datum.Time))
			Expect(document["origin"]).To(BeAssignableToTypeOf(map[string]any{}))
			Expect(document["tags"]).To(BeAssignableToTypeOf([]any{}))
			Expect(document["revisions"]).To(BeAssignableToTypeOf([]any{}))
			Expect(document["revisions"].([]any)[0]).To(BeAssignableToTypeOf(map[string]any{}))
			Expect(document["rate"]).To(Equal(1.25))
			Expect(document["_userId"]).To(Equal(*datum.UserID))
			Expect(document["_version"]).To(Equal(int32(3)))
			Expect(document).ToNot(HaveKey("createdUserID"))
		})
	})

	Context("Project", func() {
		It("does not modify the document", func() {
			document := map[string]any{"_id": "x", "time": "t"}
			revision.Project(document)
			Expect(document).To(HaveLen(2))
		})
	})

	Context("Canonicalize", func() {
		It("drops null members recursively and keeps nulls inside arrays", func() {
			Expect(revision.Canonicalize(map[string]any{
				"a": nil,
				"b": map[string]any{"c": nil, "d": "e"},
				"f": []any{nil, map[string]any{"g": nil, "h": 1}},
			})).To(Equal(map[string]any{
				"b": map[string]any{"d": "e"},
				"f": []any{nil, map[string]any{"h": int64(1)}},
			}))
		})

		It("replaces BSON map and array types with plain maps and slices", func() {
			Expect(revision.Canonicalize(map[string]any{
				"m": primitive.M{"x": primitive.A{int32(1), primitive.D{{Key: "y", Value: "z"}}}},
			})).To(Equal(map[string]any{
				"m": map[string]any{"x": []any{int64(1), map[string]any{"y": "z"}}},
			}))
		})

		It("widens every integer type to int64 and float32 to float64", func() {
			Expect(revision.Canonicalize(map[string]any{
				"a": int(1), "b": int8(2), "c": int16(3), "d": int32(4), "e": int64(5),
				"f": uint(6), "g": uint8(7), "h": uint16(8), "i": uint32(9), "j": uint64(10),
				"k": float32(1.5), "l": 2.5,
			})).To(Equal(map[string]any{
				"a": int64(1), "b": int64(2), "c": int64(3), "d": int64(4), "e": int64(5),
				"f": int64(6), "g": int64(7), "h": int64(8), "i": int64(9), "j": int64(10),
				"k": float64(1.5), "l": 2.5,
			}))
		})

		It("converts BSON datetimes and local times to UTC times", func() {
			tm := randomTime()
			local := tm.In(time.FixedZone("plus", 3600))
			result := revision.Canonicalize(map[string]any{"a": primitive.NewDateTimeFromTime(tm), "b": local})
			Expect(result["a"]).To(Equal(tm))
			Expect(result["b"]).To(Equal(tm))
			Expect(result["b"].(time.Time).Location()).To(Equal(time.UTC))
		})

		It("keeps strings, booleans and object ids", func() {
			id := primitive.NewObjectID()
			Expect(revision.Canonicalize(map[string]any{"a": "s", "b": true, "c": id})).To(Equal(map[string]any{"a": "s", "b": true, "c": id}))
		})

		It("returns a deep copy", func() {
			document := map[string]any{"a": map[string]any{"b": []any{"c"}}}
			result := revision.Canonicalize(document)
			result["a"].(map[string]any)["b"].([]any)[0] = "changed"
			Expect(document).To(Equal(map[string]any{"a": map[string]any{"b": []any{"c"}}}))
		})
	})

	Context("Diff and CreatePatch", func() {
		DescribeTable("creates the patch that merges from into to",
			func(from string, to string, expected string) {
				fromValue := mapFromJSON(from)
				toValue := mapFromJSON(to)
				patch := revision.CreatePatch(fromValue, toValue)
				Expect(patch).To(Equal(fromJSON(expected)))
				Expect(revision.Apply(fromValue, patch)).To(Equal(toValue))
			},
			Entry("identical", `{"a":"b","c":{"d":1}}`, `{"a":"b","c":{"d":1}}`, `{}`),
			Entry("scalar changed", `{"a":"b"}`, `{"a":"c"}`, `{"a":"c"}`),
			Entry("key added", `{"a":"b"}`, `{"a":"b","c":"d"}`, `{"c":"d"}`),
			Entry("key removed", `{"a":"b","c":"d"}`, `{"a":"b"}`, `{"c":null}`),
			Entry("nested changed", `{"a":{"b":"c","d":"e"}}`, `{"a":{"b":"x","d":"e"}}`, `{"a":{"b":"x"}}`),
			Entry("nested key removed", `{"a":{"b":"c","d":"e"}}`, `{"a":{"d":"e"}}`, `{"a":{"b":null}}`),
			Entry("nested object added", `{"a":"b"}`, `{"a":"b","c":{"d":{"e":1}}}`, `{"c":{"d":{"e":1}}}`),
			Entry("nested object removed", `{"a":"b","c":{"d":1}}`, `{"a":"b"}`, `{"c":null}`),
			Entry("array changed replaces whole", `{"a":[1,2,3]}`, `{"a":[1,2]}`, `{"a":[1,2]}`),
			Entry("array of objects changed replaces whole", `{"a":[{"b":1},{"c":2}]}`, `{"a":[{"b":1},{"c":3}]}`, `{"a":[{"b":1},{"c":3}]}`),
			Entry("scalar to object", `{"a":"b"}`, `{"a":{"c":"d"}}`, `{"a":{"c":"d"}}`),
			Entry("object to scalar", `{"a":{"c":"d"}}`, `{"a":"b"}`, `{"a":"b"}`),
			Entry("scalar to empty object", `{"a":"b"}`, `{"a":{}}`, `{"a":{}}`),
			Entry("empty object to scalar", `{"a":{}}`, `{"a":"b"}`, `{"a":"b"}`),
			Entry("number changed", `{"a":1}`, `{"a":2}`, `{"a":2}`),
			Entry("type changed", `{"a":1}`, `{"a":"1"}`, `{"a":"1"}`),
			Entry("boolean changed", `{"a":true}`, `{"a":false}`, `{"a":false}`),
		)

		It("treats a null member of to as absent", func() {
			Expect(revision.CreatePatch(map[string]any{"a": "b"}, map[string]any{"a": nil})).To(Equal(map[string]any{"a": nil}))
			Expect(revision.CreatePatch(map[string]any{}, map[string]any{"a": nil})).To(BeEmpty())
		})

		It("compares canonical times by instant", func() {
			tm := randomTime()
			from := revision.Canonicalize(map[string]any{"time": tm})
			to := revision.Canonicalize(map[string]any{"time": tm.In(time.FixedZone("plus", 3600))})
			Expect(revision.CreatePatch(from, to)).To(BeEmpty())
			later := tm.Add(time.Second)
			Expect(revision.CreatePatch(from, map[string]any{"time": later})).To(Equal(map[string]any{"time": later}))
		})

		It("compares int32 and int64 by value after canonicalization", func() {
			from := revision.Canonicalize(map[string]any{"duration": int32(3600000)})
			to := revision.Canonicalize(map[string]any{"duration": int64(3600000)})
			Expect(revision.CreatePatch(from, to)).To(BeEmpty())
		})

		It("does not compare an integer and a double as equal", func() {
			Expect(revision.CreatePatch(map[string]any{"a": int64(1)}, map[string]any{"a": float64(1)})).To(Equal(map[string]any{"a": float64(1)}))
		})

		It("compares values of uncomparable types", func() {
			a := primitive.Binary{Subtype: 0, Data: []byte{1, 2}}
			b := primitive.Binary{Subtype: 0, Data: []byte{1, 3}}
			Expect(revision.CreatePatch(map[string]any{"a": a}, map[string]any{"a": a})).To(BeEmpty())
			Expect(revision.CreatePatch(map[string]any{"a": a}, map[string]any{"a": b})).To(Equal(map[string]any{"a": b}))
		})

		It("returns a patch whose values are copies", func() {
			to := map[string]any{"a": map[string]any{"b": []any{"c"}}}
			patch := revision.CreatePatch(map[string]any{}, to)
			patch["a"].(map[string]any)["b"].([]any)[0] = "mutated"
			Expect(to).To(Equal(map[string]any{"a": map[string]any{"b": []any{"c"}}}))
		})
	})

	Context("with stored datums", func() {
		It("diffs and applies a suspend basal with a nested suppressed basal as plain maps", func() {
			datum := dataTypesBasalSuspend.New()
			datum.Base = *dataTypesTest.RandomBase()
			datum.Duration = pointer.FromInt(600000)
			datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
			originalRate := *datum.Suppressed.(*dataTypesBasalScheduled.SuppressedScheduled).Rate
			document, err := revision.ToMap(datum)
			Expect(err).ToNot(HaveOccurred())
			stored := revision.Canonicalize(revision.Project(document))
			Expect(stored["suppressed"]).To(BeAssignableToTypeOf(map[string]any{}))

			datum.Duration = pointer.FromInt(900000)
			datum.Suppressed.(*dataTypesBasalScheduled.SuppressedScheduled).Rate = pointer.FromFloat64(9.5)
			document, err = revision.ToMap(datum)
			Expect(err).ToNot(HaveOccurred())
			incoming := revision.Canonicalize(revision.Project(document))

			forward := revision.CreatePatch(stored, incoming)
			Expect(forward).To(Equal(map[string]any{"duration": int64(900000), "suppressed": map[string]any{"rate": 9.5}}))
			undo := revision.CreatePatch(incoming, stored)
			Expect(undo).To(Equal(map[string]any{"duration": int64(600000), "suppressed": map[string]any{"rate": originalRate}}))
			Expect(revision.Apply(stored, forward)).To(Equal(incoming))
			Expect(revision.Apply(incoming, undo)).To(Equal(stored))
		})

		It("diffs a calculator by its stored bolus id and an alarm by its stored status id", func() {
			calculator := dataTypesCalculator.New()
			calculator.Base = *dataTypesTest.RandomBase()
			calculator.BolusID = pointer.FromString(dataTest.RandomDatumID())
			document, err := revision.ToMap(calculator)
			Expect(err).ToNot(HaveOccurred())
			stored := revision.Canonicalize(revision.Project(document))
			Expect(stored["bolus"]).To(Equal(*calculator.BolusID))

			calculator.BolusID = pointer.FromString(dataTest.RandomDatumID())
			document, err = revision.ToMap(calculator)
			Expect(err).ToNot(HaveOccurred())
			incoming := revision.Canonicalize(revision.Project(document))
			Expect(revision.CreatePatch(stored, incoming)).To(Equal(map[string]any{"bolus": *calculator.BolusID}))

			alarm := dataTypesDeviceAlarm.New()
			alarm.Base = *dataTypesTest.RandomBase()
			alarm.AlarmType = pointer.FromString(dataTypesDeviceAlarm.AlarmTypeOcclusion)
			alarm.StatusID = pointer.FromString(dataTest.RandomDatumID())
			document, err = revision.ToMap(alarm)
			Expect(err).ToNot(HaveOccurred())
			stored = revision.Canonicalize(revision.Project(document))
			Expect(stored["status"]).To(Equal(*alarm.StatusID))
			Expect(revision.CreatePatch(stored, stored)).To(BeEmpty())
		})

		It("renders an undo to JSON exactly as the stored datum renders its own fields", func() {
			datum, stored := storedScheduled()
			incoming := revision.Apply(stored, map[string]any{"time": datum.Time.Add(time.Hour), "rate": 1.1, "duration": int64(1)})
			undo := revision.CreatePatch(incoming, stored)
			Expect(undo).To(HaveLen(3))

			undoJSON := map[string]any{}
			bytes, err := json.Marshal(undo)
			Expect(err).ToNot(HaveOccurred())
			Expect(json.Unmarshal(bytes, &undoJSON)).To(Succeed())
			datumJSON := map[string]any{}
			bytes, err = json.Marshal(datum)
			Expect(err).ToNot(HaveOccurred())
			Expect(json.Unmarshal(bytes, &datumJSON)).To(Succeed())
			for _, key := range []string{"time", "rate", "duration"} {
				Expect(undoJSON[key]).To(Equal(datumJSON[key]), key)
			}
		})

		It("strips null members for comparison and keeps nulls inside an undo", func() {
			_, stored := storedScheduled()
			withNull := revision.Apply(stored, map[string]any{"scheduleName": nil})
			withNull["payload"] = map[string]any{"a": "b", "c": nil}
			incoming := revision.Canonicalize(withNull)
			Expect(incoming).ToNot(HaveKey("scheduleName"))
			Expect(incoming["payload"]).ToNot(HaveKey("c"))

			forward := revision.CreatePatch(stored, incoming)
			Expect(forward["scheduleName"]).To(BeNil())
			Expect(forward).To(HaveKey("scheduleName"))
			undo := revision.CreatePatch(incoming, stored)
			Expect(undo["scheduleName"]).To(Equal("Weekday"))
			Expect(revision.Apply(incoming, undo)).To(Equal(stored))
		})

		It("computes the forward and undo patches independently", func() {
			_, stored := storedScheduled()
			incoming := revision.Apply(stored, map[string]any{"rate": 2.0, "scheduleName": nil, "notes": []any{"x"}})
			forward := revision.CreatePatch(stored, incoming)
			undo := revision.CreatePatch(incoming, stored)
			Expect(forward).To(Equal(map[string]any{"rate": 2.0, "scheduleName": nil, "notes": []any{"x"}}))
			Expect(undo).To(Equal(map[string]any{"rate": 0.85, "scheduleName": "Weekday", "notes": stored["notes"]}))
			Expect(revision.Apply(stored, forward)).To(Equal(incoming))
			Expect(revision.Apply(incoming, undo)).To(Equal(stored))
		})

		It("reproduces the original by applying three undos newest first", func() {
			_, original := storedScheduled()
			first := revision.Apply(original, map[string]any{"rate": 1.0})
			second := revision.Apply(first, map[string]any{"duration": int64(1800000), "scheduleName": nil})
			third := revision.Apply(second, map[string]any{"payload": map[string]any{"added": true}, "rate": 1.5})
			undos := []map[string]any{
				revision.CreatePatch(first, original),
				revision.CreatePatch(second, first),
				revision.CreatePatch(third, second),
			}
			current := third
			for index := len(undos) - 1; index >= 0; index-- {
				current = revision.Apply(current, undos[index])
			}
			Expect(current).To(Equal(original))
			Expect(revision.CreatePatch(current, original)).To(BeEmpty())
		})
	})

	Context("NewUpdate", func() {
		It("sets whole values from the result for changed keys and unsets removed keys", func() {
			_, stored := storedScheduled()
			incoming := revision.Apply(stored, map[string]any{"rate": 2.0, "scheduleName": nil, "origin": map[string]any{"id": "new"}})
			patch := revision.CreatePatch(stored, incoming)
			update, err := revision.NewUpdate(patch, incoming)
			Expect(err).ToNot(HaveOccurred())
			Expect(patch["origin"]).To(Equal(map[string]any{"id": "new"}), "the patch carries only the changed member")
			Expect(update).To(Equal(bson.M{
				"$set":   bson.M{"rate": 2.0, "origin": incoming["origin"]},
				"$unset": bson.M{"scheduleName": 1},
			}), "the update sets the whole top-level value")
			Expect(update["$set"].(bson.M)["origin"]).To(HaveLen(len(stored["origin"].(map[string]any))))
		})

		It("returns an empty update for an empty patch", func() {
			update, err := revision.NewUpdate(map[string]any{}, map[string]any{})
			Expect(err).ToNot(HaveOccurred())
			Expect(update).To(BeEmpty())
		})

		It("refuses to set a server-owned key", func() {
			update, err := revision.NewUpdate(map[string]any{"_version": int64(2)}, map[string]any{"_version": int64(2)})
			Expect(err).To(MatchError(`key "_version" is server owned`))
			Expect(update).To(BeNil())
		})

		It("refuses to unset a server-owned key", func() {
			update, err := revision.NewUpdate(map[string]any{"revisions": nil}, map[string]any{})
			Expect(err).To(MatchError(`key "revisions" is server owned`))
			Expect(update).To(BeNil())
		})

		It("refuses a set key that is missing from the result", func() {
			update, err := revision.NewUpdate(map[string]any{"rate": 2.0}, map[string]any{})
			Expect(err).To(MatchError(`key "rate" is missing from result`))
			Expect(update).To(BeNil())
		})
	})

	Context("datum types", func() {
		// Clients fold an undo into the datum JSON by key, so every key inside the mutable projection must have the
		// same name in BSON and JSON. Server-owned keys are outside the projection and may differ.
		It("name every client-visible key identically in JSON and BSON", func() {
			datums := []any{
				dataTypesActivityPhysical.New(),
				dataTypesAlert.New(),
				dataTypesBasalAutomated.New(),
				&dataTypesBasalAutomated.SuppressedAutomated{},
				dataTypesBasalScheduled.New(),
				&dataTypesBasalScheduled.SuppressedScheduled{},
				dataTypesBasalSuspend.New(),
				dataTypesBasalTemporary.New(),
				&dataTypesBasalTemporary.SuppressedTemporary{},
				dataTypesBloodGlucoseContinuous.New(),
				dataTypesBloodGlucoseSelfmonitored.New(),
				dataTypesBloodKetone.New(),
				dataTypesBolusAutomated.New(),
				dataTypesBolusCombination.New(),
				dataTypesBolusExtended.New(),
				dataTypesBolusNormal.New(),
				dataTypesCalculator.New(),
				dataTypesDeviceAlarm.New(),
				dataTypesDeviceCalibration.New(),
				dataTypesDeviceOverrideSettingsPump.New(),
				dataTypesDevicePrime.New(),
				dataTypesDeviceReservoirchange.New(),
				dataTypesDeviceStatus.New(),
				dataTypesDeviceTimechange.New(),
				dataTypesDosingDecision.New(),
				dataTypesFood.New(),
				dataTypesInsulin.New(),
				dataTypesSettingsCGM.New(),
				dataTypesSettingsController.New(),
				dataTypesSettingsPump.New(),
				dataTypesStateReported.New(),
				dataTypesStatusController.New(),
				dataTypesStatusPump.New(),
				dataTypesWater.New(),
			}
			visited := map[reflect.Type]bool{}
			var mismatches []string
			var walk func(typ reflect.Type, path string)
			walk = func(typ reflect.Type, path string) {
				for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array || typ.Kind() == reflect.Map {
					typ = typ.Elem()
				}
				if typ.Kind() != reflect.Struct || typ == reflect.TypeOf(time.Time{}) || visited[typ] {
					return
				}
				visited[typ] = true
				for index := range typ.NumField() {
					field := typ.Field(index)
					if field.Anonymous {
						walk(field.Type, path)
						continue
					}
					jsonName, _, _ := strings.Cut(field.Tag.Get("json"), ",")
					bsonName, _, _ := strings.Cut(field.Tag.Get("bson"), ",")
					if jsonName == "-" || bsonName == "-" {
						continue
					}
					if jsonName == "" {
						jsonName = field.Name
					}
					if bsonName == "" {
						bsonName = strings.ToLower(field.Name)
					}
					if path == "" && field.Tag.Get(revision.TagRevision) == revision.TagRevisionOmit {
						continue
					}
					if jsonName != bsonName {
						mismatches = append(mismatches, typ.String()+"."+field.Name+": json "+jsonName+", bson "+bsonName)
					}
					walk(field.Type, path+"/"+jsonName)
				}
			}
			for _, datum := range datums {
				walk(reflect.TypeOf(datum), "")
			}
			Expect(visited).To(HaveKey(reflect.TypeOf(types.Base{})))
			Expect(mismatches).To(BeEmpty())
		})

		It("stores the datum provenance under the BSON name the deduplicator queries", func() {
			datum := dataTypesWater.New()
			datum.Base = *dataTypesTest.RandomBase()
			datum.Revisions = &[]data.Revision{{Provenance: dataTest.RandomProvenance(), Time: randomTime()}}
			bytes, err := bson.Marshal(datum)
			Expect(err).ToNot(HaveOccurred())
			Expect(bson.Raw(bytes).Lookup("revisions", "0", "provenance", "dataRaw", "id").StringValue()).To(Equal((*datum.Revisions)[0].Provenance.DataRaw.ID))
		})
	})
})
