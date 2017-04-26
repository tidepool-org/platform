package app

import "time"

func StringAsPointer(source string) *string                 { return &source }
func StringArrayAsPointer(source []string) *[]string        { return &source }
func IntegerAsPointer(source int) *int                      { return &source }
func FloatAsPointer(source float64) *float64                { return &source }
func DurationAsPointer(source time.Duration) *time.Duration { return &source }
