package test

import "math/rand"

func MustStringArray(value []string, err error) []string {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomStringArray() []string {
	return RandomStringArrayFromRange(RandomStringArrayLengthMinimum(), RandomStringArrayLengthMaximum())
}

func RandomStringArrayFromRange(minimumLength int, maximumLength int) []string {
	if maximumLength < minimumLength {
		panic("RandomStringArrayFromRange: maximum length is not greater than or equal to minimum length")
	}
	result := make([]string, RandomIntFromRange(minimumLength, maximumLength))
	for index := range result {
		result[index] = RandomString()
	}
	return result
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
	result := make([]string, RandomIntFromRange(minimumLength, maximumLength))
	for index := range result {
		result[index] = RandomStringFromCharset(charset)
	}
	return result
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
	return 3
}

func RandomStringArrayLengthMinimum() int {
	return 1
}

func NewObjectFromStringArray(value []string, objectFormat ObjectFormat) interface{} {
	if value == nil {
		return nil
	}
	object := []interface{}{}
	for _, element := range value {
		object = append(object, NewObjectFromString(element, objectFormat))
	}
	return object
}
