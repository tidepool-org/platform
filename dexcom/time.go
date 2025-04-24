package dexcom

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
)

func ParseTime(parser structure.ObjectParser, reference string) *Time {
	serialized := parser.String(reference)
	if serialized == nil {
		return nil
	}

	tm, err := TimeFromString(*serialized)
	if err != nil {
		parser.WithReferenceErrorReporter(reference).ReportError(structureParser.ErrorValueTimeNotParsable(*serialized, time.RFC3339Nano))
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
	var zoneString *string
	if zoneMatches := zuluZoneRegexp.FindStringSubmatch(parsable); zoneMatches != nil {
		parsable = zoneMatches[1]
		zoneString = pointer.FromString(zoneMatches[2])
	} else if zoneMatches := offsetZoneRegexp.FindStringSubmatch(parsable); zoneMatches != nil {
		parsable = zoneMatches[1]
		zoneString, err = zoneOffsetFromString(zoneMatches[2])
		if err != nil {
			return nil, errors.Wrap(err, "time is not parsable")
		}
	}

	// Parse out all of the fields, be *very* lenient
	pipeline := parsePipeline{
		digitsParser(4, 4, &year),
		characterParser("-"),
		digitsParser(1, 2, &month),
		characterParser("-"),
		digitsParser(1, 2, &day),
		characterParser("T "),
		digitsParser(1, 2, &hour),
		characterParser(":"),
		digitsParser(1, 2, &minute),
		characterParser(":"),
		digitsParser(1, 2, &second),
		characterParser("."),
		nanosecondsPadder(),
		digitsParser(9, 9, &nanoseconds),
	}
	remaining, err := pipeline.Parse(parsable)

	// If we had an error or there are unparsed characters left, then bail
	if err != nil || remaining != "" {
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
	if zoneString != nil {
		parsable += *zoneString
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
		zoneNotParsed: zoneString == nil,
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
	if t == nil {
		return nil
	}
	return &t.Time
}

func (t Time) String() string {
	if t.serialized != "" {
		return t.serialized
	} else {
		return t.Format(time.RFC3339Nano)
	}
}

func (t *Time) ZoneParsed() bool {
	return !t.zoneNotParsed
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func zoneOffsetFromString(serialized string) (*string, error) {
	var direction *string
	var hour *int
	var minute *int
	var second *int
	var separator *string

	// Parse a copy to retain original
	parsable := serialized

	// Parse out all of the fields, be *very* lenient
	pipeline := parsePipeline{
		charactersParser("+-", 1, 1, &direction),
		digitsParser(1, 2, &hour),
		charactersParser(":", 0, 1, &separator),
		digitsParser(1, 2, &minute),
		charactersParser(":", 0, 1, &separator),
		digitsParser(1, 2, &second),
	}
	remaining, err := pipeline.Parse(parsable)

	// If we had an error or there are unparsed characters left, then bail
	if err != nil || remaining != "" || direction == nil || hour == nil {
		return nil, errors.New("zone is not parsable")
	}

	// Some Dexcom data has a time zone offset that includes seconds (e.g. "-06:59:58"), but
	// cannot be parsed correctly due to a bug in Go (https://github.com/golang/go/issues/68263).
	// Since this is obviously a time zone offset calculation issue with Dexcom, fix to the
	// nearest minute (which should correct the Dexcom issue).
	if minute != nil && second != nil && *second >= 30 {
		*minute += 1
		if *minute >= 60 {
			*minute -= 60
			*hour += 1
		}
	}

	return pointer.FromString(fmt.Sprintf("%s%02d:%02d", *direction, *hour, intOrDefault(minute, 0))), nil
}

var zuluZoneRegexp = regexp.MustCompile(`^(.*)(Z)$`)
var offsetZoneRegexp = regexp.MustCompile(`^(.*[T ].*)((\+|-)[0-9]{2,2}((|:)[0-9]{2,2}){0,2})$`)

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

// parseFn is a function which when called will run the parser and return the remaining characters or an error
type parserFn func(string) (string, error)

// parsePipeline is pipeline of parser functions
type parsePipeline []parserFn

// Parse executed the parse pipeline until an error is returned or no remaining characters are left
func (p parsePipeline) Parse(original string) (remaining string, err error) {
	remaining = original
	for _, f := range p {
		remaining, err = f(remaining)
		if err != nil || remaining == "" {
			break
		}
	}

	return
}

// digitsParser creates a digits parser function which can be added to a parser pipeline
func digitsParser(minimum int, maximum int, result **int) parserFn {
	return func(parsable string) (remaining string, err error) {
		*result, remaining, err = parseDigits(parsable, minimum, maximum)
		return
	}
}

// characterParser creates a character parser function which can be added to a parser pipeline
func characterParser(character string) parserFn {
	return func(parsable string) (string, error) {
		return parseCharacter(parsable, character)
	}
}

// charactersParser creates a characters parser function which can be added to a parser pipeline
func charactersParser(character string, minimum int, maximum int, result **string) parserFn {
	return func(parsable string) (remaining string, err error) {
		*result, remaining, err = parseCharacters(parsable, character, minimum, maximum)
		return
	}
}

// nanosecondsPadder pads the passed string with up to nine '0' chars if the length of string is less than nine
func nanosecondsPadder() parserFn {
	return func(parsable string) (string, error) {
		if length := len(parsable); length < 9 {
			parsable += strings.Repeat("0", 9-length)
		}
		return parsable, nil
	}
}
