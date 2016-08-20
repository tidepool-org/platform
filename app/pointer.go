package app

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import "time"

func StringAsPointer(source string) *string                 { return &source }
func StringArrayAsPointer(source []string) *[]string        { return &source }
func IntegerAsPointer(source int) *int                      { return &source }
func DurationAsPointer(source time.Duration) *time.Duration { return &source }
