package pointer

import "time"

func Boolean(source bool) *bool                    { return &source }
func Duration(source time.Duration) *time.Duration { return &source }
func Float(source float64) *float64                { return &source }
func Integer(source int) *int                      { return &source }
func String(source string) *string                 { return &source }
func StringArray(source []string) *[]string        { return &source }
