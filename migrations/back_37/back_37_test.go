package main

import (
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
)

func Test_getBGValuePlatformPrecision(t *testing.T) {

	tests := []struct {
		name             string
		mmolJellyfishVal float64
		mmolPlatformVal  float64
	}{
		{
			name:             "original mmol/L value",
			mmolJellyfishVal: 10.1,
			mmolPlatformVal:  10.1,
		},
		{
			name:             "converted mgd/L of 100",
			mmolJellyfishVal: 5.550747991045533,
			mmolPlatformVal:  5.55075,
		},
		{
			name:             "converted mgd/L of 150.0",
			mmolJellyfishVal: 8.3261219865683,
			mmolPlatformVal:  8.32612,
		},
		{
			name:             "converted mgd/L of 65.0",
			mmolJellyfishVal: 3.6079861941795968,
			mmolPlatformVal:  3.60799,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBGValuePlatformPrecision(tt.mmolJellyfishVal); got != tt.mmolPlatformVal {
				t.Errorf("getBGValuePlatformPrecision() mmolJellyfishVal = %v, want %v", got, tt.mmolPlatformVal)
			}
		})
	}
}

func Test_updateIfExistsPumpSettingsBolus(t *testing.T) {
	type args struct {
		bsonData bson.M
	}

	bolusData := map[string]interface{}{
		"bolous-1": pumpTest.NewRandomBolus(),
		"bolous-2": pumpTest.NewRandomBolus(),
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "when not pumpSettings",
			args: args{
				bsonData: bson.M{"type": "other"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "pumpSettings but no bolus",
			args: args{
				bsonData: bson.M{"type": "pumpSettings"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "pumpSettings bolus wrong type",
			args: args{
				bsonData: bson.M{"type": "pumpSettings", "bolus": "wrong"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "pumpSettings bolus valid type",
			args: args{
				bsonData: bson.M{"type": "pumpSettings", "bolus": bolusData},
			},
			want:    bolusData,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateIfExistsPumpSettingsBolus(tt.args.bsonData)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateIfExistsPumpSettingsBolus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateIfExistsPumpSettingsBolus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateIfExistsPumpSettingsSleepSchedules(t *testing.T) {

	type args struct {
		bsonData bson.M
	}

	sleepSchedulesExpected := &pump.SleepScheduleMap{
		"schedule-1": pumpTest.RandomSleepSchedule(),
		"schedule-2": pumpTest.RandomSleepSchedule(),
	}
	sleepSchedulesStored := pumpTest.CloneSleepSchedules(sleepSchedulesExpected)
	sleepSchedulesInvalidDays := pumpTest.CloneSleepSchedules(sleepSchedulesExpected)

	//ensure sorting
	sleepSchedulesExpected.Normalize(normalizer.New())

	s1Days := (*sleepSchedulesStored)["schedule-1"].Days
	for key, day := range *s1Days {
		(*s1Days)[key] = strings.ToUpper(day)
	}
	(*sleepSchedulesStored)["schedule-1"].Days = s1Days

	s2Days := (*sleepSchedulesStored)["schedule-2"].Days

	for key, day := range *s2Days {
		(*s2Days)[key] = strings.ToUpper(day)
	}
	(*sleepSchedulesStored)["schedule-2"].Days = s2Days

	(*sleepSchedulesInvalidDays)["schedule-2"].Days = &[]string{"not-a-day", common.DayFriday}

	tests := []struct {
		name    string
		args    args
		want    *pump.SleepScheduleMap
		wantErr bool
	}{
		{
			name: "when not pumpSettings type",
			args: args{
				bsonData: bson.M{"type": "other"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "pumpSettings but no sleepSchedules",
			args: args{
				bsonData: bson.M{"type": "pumpSettings"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "pumpSettings sleepSchedules wrong type",
			args: args{
				bsonData: bson.M{"type": "pumpSettings", "sleepSchedules": "wrong"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "pumpSettings sleepSchedules days invalid",
			args: args{
				bsonData: bson.M{"type": "pumpSettings", "sleepSchedules": sleepSchedulesInvalidDays},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "pumpSettings sleepSchedules valid type",
			args: args{
				bsonData: bson.M{"type": "pumpSettings", "sleepSchedules": sleepSchedulesStored},
			},
			want:    sleepSchedulesExpected,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateIfExistsPumpSettingsSleepSchedules(tt.args.bsonData)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateIfExistsPumpSettingsSleepSchedules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				gotSleepSchedules, ok := got.(*pump.SleepScheduleMap)
				if !ok {
					t.Errorf("updateIfExistsPumpSettingsSleepSchedules() = %v, want %v", got, tt.want)
					return
				}

				for key, wantSchedule := range *tt.want {
					if gotSchedule := (*gotSleepSchedules)[key]; gotSchedule != nil {
						if !reflect.DeepEqual((*gotSchedule).Days, (*wantSchedule).Days) {
							t.Errorf("updateIfExistsPumpSettingsSleepSchedules() = %v, want %v", (*gotSchedule).Days, (*wantSchedule).Days)
						}
					} else {
						t.Errorf("missing schedule %s", key)
					}
				}
			}
		})
	}
}
