package pointer

import "time"

func Bool(source bool) *bool                       { return &source }
func Duration(source time.Duration) *time.Duration { return &source }
func Float64(source float64) *float64              { return &source }
func Int(source int) *int                          { return &source }
func String(source string) *string                 { return &source }
func StringArray(source []string) *[]string        { return &source }
func Time(source time.Time) *time.Time             { return &source }
