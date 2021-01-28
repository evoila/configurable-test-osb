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
	/*if invalidFrequency(&settings.OfferingBindable) {
		return errors.New("invalid value for offering_bindable")
	}
	if invalidFrequency(&settings.InstancesRetrievable) {
		return errors.New("invalid value for instances_retrievable")
	}
	if invalidFrequency(&settings.BindingsRetrievable) {
		return errors.New("invalid value for bindings_retrievable")
	}
	if invalidFrequency(&settings.AllowContextUpdates) {
		return errors.New("invalid value for allow_context_updates")
	}
	if invalidFrequency(&settings.OfferingMetadata) {
		return errors.New("invalid value for offering_metadata")
	}
	if invalidFrequency(&settings.DashboardClient) {
		return errors.New("invalid value for dashboard_client")
	}
	if invalidFrequency(&settings.OfferingPlanUpdateable) {
		return errors.New("invalid value for offering_plan_updateable")
	}
	if settings.PlansMin < 1 {
		return errors.New("plans_min must be > 0")
	}
	if settings.PlansMax < settings.PlansMin {
		return errors.New("plans_max must be >= plans_min")
	}
	if invalidFrequency(&settings.PlanMetadata) {
		return errors.New("invalid value for plan_metadata")
	}
	if invalidFrequency(&settings.Free) {
		return errors.New("invalid value for free")
	}
	if invalidFrequency(&settings.PlanBindable) {
		return errors.New("invalid value for plan_bindable")
	}
	if invalidFrequency(&settings.BindingRotatable) {
		return errors.New("invalid value for binding_rotatable")
	}
	*/
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
	fmt.Println("Calling invalidRequires")
	fmt.Println(requires)
	for _, val := range requires {
		fmt.Println(val)
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

/*
func init() {
	//WILL PROBABLY NOT WORK BECAUSE GIN WON'T BE USE FOR THIS STRUCT
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("RFC3986", validateRfc3986)
		if err != nil {
			log.Println("Error while registering rfc3986Validator to validations! error: " + err.Error())
		}
	}

}

*/
