package dexcom

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
)

func ParseTime(reference string, parser structure.ObjectParser) *Time {
	serialized := parser.String(reference)
	if serialized == nil {
		return nil
	}

	tm, err := TimeFromString(*serialized)
	if err != nil {
		parser.ReportError(structureParser.ErrorValueTimeNotParsable(*serialized, time.RFC3339Nano))
		return nil
	}

	return tm
}

// HACK: Dexcom V3 (2024-05-30) - Times may not include seconds (e.g. "2021-02-03T14:15Z")
// Assume this applies to all fields and separators being optional plus any single digit
// may not be prefixed with a zero.
func TimeFromString(serialized string) (*Time, error) {
	var year *int
	var month *int
	var day *int
	var hour *int
	var minute *int
	var second *int
	var nanoseconds *int
	var err error

	// Parse a copy to retain original
	parsable := serialized

	// Determine if there is a zone, if so only use the non-zone portion
	zoneMatches := zoneRegexp.FindStringSubmatch(parsable)
	if zoneMatches != nil {
		parsable = zoneMatches[1]
	}

	// Parse out all of the fields, be *very* lenient
	if year, parsable, err = parseDigits(parsable, 4, 4); err == nil && parsable != "" {
		if parsable, err = parseCharacter(parsable, "-"); err == nil && parsable != "" {
			if month, parsable, err = parseDigits(parsable, 1, 2); err == nil && parsable != "" {
				if parsable, err = parseCharacter(parsable, "-"); err == nil && parsable != "" {
					if day, parsable, err = parseDigits(parsable, 1, 2); err == nil && parsable != "" {
						if parsable, err = parseCharacter(parsable, "T"); err == nil && parsable != "" {
							if hour, parsable, err = parseDigits(parsable, 1, 2); err == nil && parsable != "" {
								if parsable, err = parseCharacter(parsable, ":"); err == nil && parsable != "" {
									if minute, parsable, err = parseDigits(parsable, 1, 2); err == nil && parsable != "" {
										if parsable, err = parseCharacter(parsable, ":"); err == nil && parsable != "" {
											if second, parsable, err = parseDigits(parsable, 1, 2); err == nil && parsable != "" {
												if parsable, err = parseCharacter(parsable, "."); err == nil && parsable != "" {
													if length := len(parsable); length < 9 {
														parsable += strings.Repeat("0", 9-length)
													}
													if nanoseconds, parsable, err = parseDigits(parsable, 9, 9); err == nil && parsable != "" {
														return nil, errors.New("time is not parsable")
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// If we had an error, then bail
	if err != nil {
		return nil, errors.New("time is not parsable")
	}

	// Rebuild a truly parsable string
	parsable = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d",
		intOrDefault(year, 2001),
		intOrDefault(month, 1),
		intOrDefault(day, 1),
		intOrDefault(hour, 0),
		intOrDefault(minute, 0),
		intOrDefault(second, 0),
	)
	if nanoseconds != nil && *nanoseconds != 0 {
		parsable = fmt.Sprintf("%s.%09d", parsable, *nanoseconds)
		parsable = strings.TrimRight(parsable, "0") // Remove extraneous
	}

	// Add the zone portion back, if no zone then use UTC for parsing
	if zoneMatches != nil {
		parsable += zoneMatches[2]
	} else {
		parsable += "Z"
	}

	// Attempt to parse, if error, then report
	tm, err := time.Parse(time.RFC3339Nano, parsable)
	if err != nil {
		return nil, errors.New("time is not parsable")
	}

	return &Time{
		Time:          tm,
		serialized:    serialized,
		zoneNotParsed: zoneMatches == nil,
	}, nil
}

func TimeFromRaw(raw time.Time) *Time {
	return &Time{
		Time:       raw,
		serialized: raw.Format(time.RFC3339Nano),
	}
}

func TimeFromTime(tm *Time) *Time {
	if tm == nil {
		return nil
	}
	return &Time{
		Time:          tm.Time,
		serialized:    tm.serialized,
		zoneNotParsed: tm.zoneNotParsed,
	}
}

type Time struct {
	time.Time
	serialized    string
	zoneNotParsed bool // Negated so the default of false is correct if struct manually created
}

func (t *Time) Raw() *time.Time {
	if t == nil || t.IsZero() {
		return nil
	}
	return &t.Time
}

func (t *Time) ZoneParsed() bool {
	return !t.zoneNotParsed
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.serialized != "" {
		return json.Marshal(t.serialized)
	} else {
		return json.Marshal(t.Format(time.RFC3339Nano))
	}
}

var zoneRegexp = regexp.MustCompile(`^(.*)(Z|-\d\d:\d\d)$`)

func parseDigits(original string, minimum int, maximum int) (*int, string, error) {
	digits, remaining, err := parseCharacters(original, "1234567890", minimum, maximum)
	if err != nil {
		return nil, original, err
	}
	number, err := strconv.ParseInt(*digits, 10, 64)
	if err != nil {
		return nil, original, errors.Wrap(err, "string is not a valid number")
	}
	return pointer.FromInt(int(number)), remaining, nil
}

func parseCharacter(original string, character string) (string, error) {
	_, remaining, err := parseCharacters(original, character, 1, 1)
	return remaining, err
}

func parseCharacters(original string, characters string, minimum int, maximum int) (*string, string, error) {
	if minimum < 0 {
		return nil, original, errors.New("minimum is less than zero")
	} else if maximum < minimum {
		return nil, original, errors.New("maximum is less than minimum")
	}

	if maximum == 0 {
		return pointer.FromString(""), original, nil
	}

	var ruins []rune
	for _, ruin := range original {
		if !strings.ContainsRune(characters, ruin) {
			break
		}
		ruins = append(ruins, ruin)
		if len(ruins) >= maximum {
			break
		}
	}
	if len(ruins) < minimum {
		return nil, original, errors.New("string does not contain minimum number of characters")
	}
	return pointer.FromString(string(ruins)), original[len(ruins):], nil
}

func intOrDefault(value *int, defowlt int) int {
	if value != nil {
		return *value
	} else {
		return defowlt
	}
}
