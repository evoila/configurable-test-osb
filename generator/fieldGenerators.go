package generator

import (
	"math/rand"
	"strconv"
)

func RandomString(n int) string {
	const characters = "abcdefghijklmnopqrstuvxyz0123456789"
	randomCharSequence := make([]byte, n)
	for i := range randomCharSequence {
		randomCharSequence[i] = characters[rand.Int63()%int64(len(characters))]
	}
	return string(randomCharSequence)
}

func ReturnBoolean(frequency string) *bool {
	booleanValue := false
	if frequency == "always" {
		booleanValue = true
		return &booleanValue
	}
	if frequency == "random" {
		if rand.Intn(2) == 1 {
			booleanValue = true
			return &booleanValue
		}
	}
	return &booleanValue
}
func ReturnFieldByBoolean(boolean *bool, frequency string) *bool {
	if *boolean {
		return ReturnBoolean(frequency)
	}
	return nil
}

func ContainsString(strings []string, element string) bool {
	for _, val := range strings {
		if val == element {
			return true
		}
	}
	return false
}

func SelectRandomTags(tags []string, min int, max int) []string {
	amount := rand.Intn(max+1-min) + min
	var result []string
	for i := 0; i < amount; i++ {
		tag := tags[rand.Int63()%int64(len(tags))]
		if ContainsString(result, tag) {
			i--
		} else {
			result = append(result, tag)
		}
	}
	return result
}

func RandomRequires(requires []string, min int) []string {
	amount := rand.Intn(len(requires)+1-min) + min
	var result []string
	for i := 0; i < amount; i++ {
		value := requires[rand.Int63()%int64(len(requires))]
		if ContainsString(result, value) {
			i--
		} else {
			result = append(result, value)
		}
	}
	return result

}

func MetadataByBool(b *bool) interface{} {
	if *b {
		return "metadata"
	}
	return nil
}

func RandomUriByFrequency(frequency string, length int) *string {
	var result string
	if *ReturnBoolean(frequency) {
		result = "http://" + RandomString(length) + ":" + strconv.Itoa(rand.Intn(9999+1-80)+80)
	}
	return &result
}
