package test

import (
	"math"
	"math/rand"
	"time"
)

func RandomBool() bool {
	return rand.Intn(2) == 0
}

func RandomBytes() []byte {
	return RandomBytesFromRange(RandomBytesLengthMinimum(), RandomBytesLengthMaximum())
}

func RandomBytesFromRange(minimumLength int, maximumLength int) []byte {
	if maximumLength < minimumLength {
		panic("RandomBytesFromRange: maximum length is not greater than or equal to minimum length")
	}
	bytes := make([]byte, RandomIntFromRange(minimumLength, maximumLength))
	length, err := rand.Read(bytes)
	if err != nil || length != len(bytes) {
		panic("RandomBytesFromRange: unable to read random bytes")
	}
	return bytes
}

func RandomBytesLengthMaximum() int {
	return 1024
}

func RandomBytesLengthMinimum() int {
	return 1
}

func RandomDuration() time.Duration {
	return RandomDurationFromRange(RandomDurationMinimum(), RandomDurationMaximum())
}

func RandomDurationFromArray(array []time.Duration) time.Duration {
	if len(array) == 0 {
		panic("RandomDurationFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomDurationFromRange(minimum time.Duration, maximum time.Duration) time.Duration {
	if maximum < minimum {
		panic("RandomDurationFromRange: maximum is not greater than or equal to minimum")
	}
	if minimum < RandomDurationMinimum() {
		minimum = RandomDurationMinimum()
	}
	if maximum > RandomDurationMaximum() {
		maximum = RandomDurationMaximum()
	}
	return minimum + time.Duration(rand.Int63n(int64(maximum-minimum+1)))
}

func RandomDurationMaximum() time.Duration {
	return 10 * 365 * 24 * time.Hour
}

func RandomDurationMinimum() time.Duration {
	return -10 * 365 * 24 * time.Hour
}

func RandomFloat64() float64 {
	return RandomFloat64FromRange(RandomFloat64Minimum(), RandomFloat64Maximum())
}

func RandomFloat64FromArray(array []float64) float64 {
	if len(array) == 0 {
		panic("RandomFloat64FromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomFloat64FromRange(minimum float64, maximum float64) float64 {
	if maximum < minimum {
		panic("RandomFloat64FromRange: maximum is not greater than or equal to minimum")
	}
	if minimum < RandomFloat64Minimum() {
		minimum = RandomFloat64Minimum()
	}
	if maximum > RandomFloat64Maximum() {
		maximum = RandomFloat64Maximum()
	}
	return minimum + (maximum-minimum+math.SmallestNonzeroFloat64)*rand.Float64()
}

func RandomFloat64Maximum() float64 {
	return math.MaxFloat32
}

func RandomFloat64Minimum() float64 {
	return -math.MaxFloat32
}

func RandomInt() int {
	return rand.Int()
}

func RandomIntFromArray(array []int) int {
	if len(array) == 0 {
		panic("RandomIntFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomIntFromRange(minimum int, maximum int) int {
	if maximum < minimum {
		panic("RandomIntFromRange: maximum is not greater than or equal to minimum")
	}
	if minimum < RandomIntMinimum() {
		minimum = RandomIntMinimum()
	}
	if maximum > RandomIntMaximum() {
		maximum = RandomIntMaximum()
	}
	return minimum + rand.Intn(maximum-minimum+1)
}

func RandomIntMaximum() int {
	return math.MaxInt32
}

func RandomIntMinimum() int {
	return math.MinInt32
}

func RandomString() string {
	return RandomStringFromRangeAndCharset(RandomStringLengthMinimum(), RandomStringLengthMaximum(), CharsetText)
}

func RandomStringFromArray(array []string) string {
	if len(array) == 0 {
		panic("RandomStringFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomStringFromCharset(charset string) string {
	return RandomStringFromRangeAndCharset(RandomStringLengthMinimum(), RandomStringLengthMaximum(), charset)
}

func RandomStringFromRange(minimumLength int, maximumLength int) string {
	return RandomStringFromRangeAndCharset(minimumLength, maximumLength, CharsetText)
}

func RandomStringFromRangeAndCharset(minimumLength int, maximumLength int, charset string) string {
	if maximumLength < minimumLength {
		panic("RandomStringFromRangeAndCharset: maximum length is not greater than or equal to minimum length")
	}
	if len(charset) == 0 {
		panic("RandomStringFromRangeAndCharset: charset is empty")
	}
	charsetRunes := []rune(charset)
	runes := make([]rune, RandomIntFromRange(minimumLength, maximumLength))
	for index := range runes {
		runes[index] = charsetRunes[rand.Intn(len(charsetRunes))]
	}
	return string(runes)
}

func RandomStringLengthMaximum() int {
	return 64
}

func RandomStringLengthMinimum() int {
	return 1
}

func RandomStringArray() []string {
	return RandomStringArrayFromRange(RandomStringArrayLengthMinimum(), RandomStringArrayLengthMaximum())
}

func RandomStringArrayFromRange(minimumLength int, maximumLength int) []string {
	if maximumLength < minimumLength {
		panic("RandomStringArrayFromRange: maximum length is not greater than or equal to minimum length")
	}
	array := make([]string, RandomIntFromRange(minimumLength, maximumLength))
	for index := range array {
		array[index] = RandomString()
	}
	return array
}

func RandomStringArrayFromRangeAndArrayWithDuplicates(minimumLength int, maximumLength int, array []string) []string {
	if maximumLength < minimumLength {
		panic("RandomStringArrayFromRangeAndArrayWithDuplicates: maximum length is not greater than or equal to minimum length")
	}
	if len(array) == 0 {
		panic("RandomStringArrayFromRangeAndArrayWithDuplicates: array is empty")
	}
	result := make([]string, RandomIntFromRange(minimumLength, maximumLength))
	for index := range result {
		result[index] = RandomStringFromArray(array)
	}
	return result
}

func RandomStringArrayFromRangeAndArrayWithoutDuplicates(minimumLength int, maximumLength int, array []string) []string {
	if maximumLength < minimumLength {
		panic("RandomStringArrayFromRangeAndArrayWithoutDuplicates: maximum length is not greater than or equal to minimum length")
	}
	if len(array) == 0 {
		panic("RandomStringArrayFromRangeAndArrayWithoutDuplicates: array is empty")
	}
	if maximumLength > len(array) {
		panic("RandomStringArrayFromRangeAndArrayWithoutDuplicates: maximum length is not less than or equal to array length")
	}
	arrayIndexes := rand.Perm(len(array))
	result := make([]string, RandomIntFromRange(minimumLength, maximumLength))
	for index := range result {
		result[index] = array[arrayIndexes[index]]
	}
	return result
}

func RandomStringArrayFromRangeAndCharset(minimumLength int, maximumLength int, charset string) []string {
	if maximumLength < minimumLength {
		panic("RandomStringArrayFromRangeAndCharset: maximum length is not greater than or equal to minimum length")
	}
	if len(charset) == 0 {
		panic("RandomStringArrayFromRangeAndCharset: charset is empty")
	}
	array := make([]string, RandomIntFromRange(minimumLength, maximumLength))
	for index := range array {
		array[index] = RandomStringFromCharset(charset)
	}
	return array
}

func RandomStringArrayFromRangeAndGeneratorWithDuplicates(minimumLength int, maximumLength int, generator func() string) []string {
	if maximumLength < minimumLength {
		panic("RandomStringArrayFromRangeAndGeneratorWithDuplicates: maximum length is not greater than or equal to minimum length")
	}
	if generator == nil {
		panic("RandomStringArrayFromRangeAndGeneratorWithDuplicates: generator is missing")
	}
	result := make([]string, RandomIntFromRange(minimumLength, maximumLength))
	for index := range result {
		result[index] = generator()
	}
	return result
}

func RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(minimumLength int, maximumLength int, generator func() string) []string {
	if maximumLength < minimumLength {
		panic("RandomStringArrayFromRangeAndGeneratorWithoutDuplicates: maximum length is not greater than or equal to minimum length")
	}
	if generator == nil {
		panic("RandomStringArrayFromRangeAndGeneratorWithoutDuplicates: generator is missing")
	}
	var result []string
	exists := map[string]bool{}
	for length := RandomIntFromRange(minimumLength, maximumLength); len(result) < length; {
		if generated := generator(); !exists[generated] {
			result = append(result, generated)
			exists[generated] = true
		}
	}
	return result
}

func RandomStringArrayLengthMaximum() int {
	return 8
}

func RandomStringArrayLengthMinimum() int {
	return 1
}

func RandomTime() time.Time {
	return RandomTimeFromRange(RandomTimeMinimum(), RandomTimeMaximum())
}

func RandomTimeFromArray(array []time.Time) time.Time {
	if len(array) == 0 {
		panic("RandomTimeFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomTimeFromRange(minimum time.Time, maximum time.Time) time.Time {
	if maximum.Before(minimum) {
		panic("RandomTimeFromRange: maximum is not greater than or equal to minimum")
	}
	if minimum.Before(RandomTimeMinimum()) {
		minimum = RandomTimeMinimum()
	}
	if maximum.After(RandomTimeMaximum()) {
		maximum = RandomTimeMaximum()
	}
	return minimum.Add(time.Duration(rand.Int63n(int64(maximum.Sub(minimum))))).Truncate(time.Millisecond)
}

func RandomTimeMaximum() time.Time {
	return now.Add(RandomDurationMaximum()).Truncate(time.Millisecond)
}

func RandomTimeMinimum() time.Time {
	return now.Add(RandomDurationMinimum()).Truncate(time.Millisecond)
}
