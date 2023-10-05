package main

import "testing"

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
