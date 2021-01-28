package generator

import (
	"math/rand"
	"strconv"
)

func randomString(n int) string {
	const characters = "abcdefghijklmnopqrstuvxyz0123456789"
	randomCharSequence := make([]byte, n)
	for i := range randomCharSequence {
		randomCharSequence[i] = characters[rand.Int63()%int64(len(characters))]
	}
	return string(randomCharSequence)
}

func returnBoolean(frequency string) *bool {
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
func returnFieldByBoolean(boolean *bool, frequency string) *bool {
	if *boolean {
		return returnBoolean(frequency)
	}
	return nil
}

func containsString(strings []string, element string) bool {
	for _, val := range strings {
		if val == element {
			return true
		}
	}
	return false
}

func selectRandomTags(tags []string, min int, max int) []string {
	amount := rand.Intn(max+1-min) + min
	var result []string
	for i := 0; i < amount; i++ {
		tag := tags[rand.Int63()%int64(len(tags))]
		if containsString(result, tag) {
			i--
		} else {
			result = append(result, tag)
		}
	}
	return result
}

func randomRequires(requires []string, min int) []string {
	//requires := [3]string{"syslog_drain", "route_forwarding", "volume_mount"}
	amount := rand.Intn(len(requires)+1-min) + min
	var result []string
	for i := 0; i < amount; i++ {
		value := requires[rand.Int63()%int64(len(requires))]
		if containsString(result, value) {
			i--
		} else {
			result = append(result, value)
		}
	}
	return result

}

func metadataByBool(b *bool) interface{} {
	if *b {
		return "metadata"
	}
	return nil
}

func randomUriByFrequency(frequency string, length int) *string {
	var result string
	if *returnBoolean(frequency) {
		result = "http://" + randomString(length) + ":" + strconv.Itoa(rand.Intn(9999+1-80)+80)
	}
	return &result
}
