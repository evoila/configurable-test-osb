package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
)

type CatalogSettings struct {
	//These are the config for service offerings and the fields it uses

	/*
		The tag binding:"required" has no effect for decoding json files to structs but is here in case an endpoint
		will be created for providing catalogSettings in a request body
	*/
	//amount > 0
	Amount           int      `json:"amount" binding:"required"`
	TagsMin          int      `json:"tags_min" binding:"required"`
	TagsMax          int      `json:"tags_max" binding:"required"`
	Requires         []string `json:"requires" binding:"required"`
	RequiresMin      int      `json:"requires_min" binding:"required"`
	OfferingBindable string   `json:"offering_bindable" binding:"required"`

	//NEW
	InstancesRetrievableExists string `json:"instance_retrievable_exists" binding:"required"`
	//
	InstancesRetrievable string `json:"instances_retrievable" binding:"required"`
	//NEW
	BindingsRetrievableExists string `json:"bindings_retrievable_exists" binding:"required"`
	//
	BindingsRetrievable string `json:"bindings_retrievable" binding:"required"`
	//NEW
	AllowContextUpdatesExists string `json:"allow_context_updates_exists" binding:"required"`
	//
	AllowContextUpdates string `json:"allow_context_updates" binding:"required"`
	OfferingMetadata    string `json:"offering_metadata" binding:"required"`
	DashboardClient     string `json:"dashboard_client" binding:"required"`
	//NEW
	OfferingPlanUpdateableExists string `json:"offering_plan_updateable_exists" binding:"required"`
	//
	OfferingPlanUpdateable string `json:"offering_plan_updateable" binding:"required"`
	//PlansMin > 0
	PlansMin int `json:"plans_min" binding:"required"`
	//PlansMax >= PlansMin
	PlansMax     int    `json:"plans_max" binding:"required"`
	PlanMetadata string `json:"plan_metadata" binding:"required"`
	//NEW
	FreeExists string `json:"free_exists" binding:"required"`
	//
	Free string `json:"free" binding:"required"`
	//NEW
	PlanBindableExists string `json:"plan_bindable_exists" binding:"required"`
	//
	PlanBindable string `json:"plan_bindable" binding:"required"`
	//NEW
	BindingRotatableExists string `json:"binding_rotatable_exists" binding:"required"`
	//
	BindingRotatable string `json:"binding_rotatable" binding:"required"`
	//NEW
	PlanUpdateableExists string `json:"plan_updateable_exists" binding:"required"`
	//
	PlanUpdateable             string `json:"plan_updateable" binding:"required"`
	Schemas                    string `json:"schemas" binding:"required"`
	ServiceInstanceSchema      string `json:"service_instance_schema" binding:"required"`
	ServiceBindingSchema       string `json:"service_binding_schema" binding:"required"`
	MaxPollingDurationMin      int    `json:"max_polling_duration_min" binding:"required"`
	MaxPollingDurationMax      int    `json:"max_polling_duration_max" binding:"required"`
	MaintenanceInfo            string `json:"maintenance_info" binding:"required"`
	MaintenanceInfoVersion     string `json:"maintenance_info_version" binding:"required"`
	MaintenanceInfoDescription string `json:"maintenance_info_description" binding:"required"`
	DashboardRedirectUri       string `json:"dashboard_redirect_uri" binding:"required"`
}

func NewCatalogSettings() (*CatalogSettings, error) {
	var catalogSettings CatalogSettings
	catalogSettingsJson, err := os.Open("config/catalogSettings.json")
	if err != nil {
		log.Println("Error while opening config/catalogSettings.json! error: " + err.Error())
		return nil, err
	}
	decoder := json.NewDecoder(catalogSettingsJson)
	if err = decoder.Decode(&catalogSettings); err != nil {
		return nil, err
	}
	if err = ValidateCatalogSettings(&catalogSettings); err != nil {
		return nil, err
	}
	log.Println("Catalog config validated!")

	s, _ := json.MarshalIndent(catalogSettings, "", "\t")
	log.Print(string(s))
	return &catalogSettings, nil
	//catalog, err := generateCatalog(&catalogSettings)
	//log.Println(catalog)

	//return catalog, err
}

func ValidateCatalogSettings(settings *CatalogSettings) error {
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
	if settings.PlansMin < 1 {
		return errors.New("plans_min must be > 0")
	}
	if settings.PlansMax < settings.PlansMin {
		return errors.New("plans_max must be >= plans_min")
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

func checkFrequencies(settings *CatalogSettings) error {
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
