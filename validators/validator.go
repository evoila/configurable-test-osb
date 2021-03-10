package validators

import (
	"errors"
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"reflect"
)

//var rfc3986Validator validator.Func =
/*
func validateRfc3986(fl validator.FieldLevel) bool {

	return true
}

*/

func ValidateCatalogSettings(settings *model.CatalogSettings) error {
	if settings.Amount < 1 {
		return errors.New("there must be at least 1 service offering")
	}
	if settings.TagsMin < 0 {
		return errors.New("tags_min min must be >= 0")
	}
	if settings.TagsMax < settings.TagsMin {
		return errors.New("tags_max must be >= tags_min")
	}
	if len(settings.Requires) > 3 {
		return errors.New("there can't be more than 3 requires as there are only 3 values")
	}
	if invalidRequires(settings.Requires) {
		return errors.New("invalid value in requires")
	}
	if numberOfDuplicates(settings.Requires) > 0 {
		return errors.New("duplicate fields in requires")
	}
	if settings.RequiresMin < 0 {
		return errors.New("requires_min min must be >= 0")
	}
	if len(settings.Requires) < settings.RequiresMin {
		return errors.New("len(requires) must be >= requires_min")
	}
	if settings.MaxPollingDurationMin < 0 {
		return errors.New("max_polling_duration_min must be >= 0")
	}
	if settings.MaxPollingDurationMax < settings.MaxPollingDurationMin {
		return errors.New("max_polling_duration_max must be >= max_polling_duration_min")
	}
	if err := checkFrequencies(settings); err != nil {
		return err
	}
	return nil
}

func checkFrequencies(settings *model.CatalogSettings) error {
	reflected := reflect.ValueOf(*settings)
	for i := 0; i < reflected.NumField(); i++ {
		value := reflected.Field(i).Interface()
		if reflect.TypeOf(value).String() == "string" {
			val := fmt.Sprintf("%s", value)
			if invalidFrequency(&val) {
				return errors.New("invalid value for " + reflected.Type().Field(i).Tag.Get("json"))
			}
		}
	}
	return nil
}

func invalidFrequency(frequency *string) bool {
	if *frequency != "never" && *frequency != "random" && *frequency != "always" {
		return true
	}
	return false
}

func invalidRequires(requires []string) bool {
	for _, val := range requires {
		if val != "syslog_drain" && val != "route_forwarding" && val != "volume_mount" {
			return true
		}
	}
	return false
}

func numberOfDuplicates(elements []string) int {
	result := 0
	for _, valA := range elements {
		for _, valB := range elements {
			if valA == valB {
				result++
			}
		}
	}
	return result - len(elements)
}
