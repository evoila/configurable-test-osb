package generator

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"io/ioutil"
	"log"
	"os"
)

type catalogSettings struct {
	//These are the settings for service offerings and the fields it uses

	//amount > 0
	Amount                 int    `json:"amount" binding:"required"`
	TagsMin                int    `json:"tags_min" binding:"required"`
	TagsMax                int    `json:"tags_max" binding:"required"`
	RequiresMin            int    `json:"requires_min" binding:"required"`
	RequiresMax            int    `json:"requires_max" binding:"required"`
	OfferingBindable       string `json:"offering_bindable" binding:"required"`
	InstancesRetrievable   string `json:"instances_retrievable" binding:"required"`
	BindingsRetrievable    string `json:"bindings_retrievable" binding:"required"`
	AllowContextUpdates    string `json:"allow_context_updates" binding:"required"`
	OfferingMetadata       string `json:"offering_metadata" binding:"required"`
	DashboardClient        string `json:"dashboard_client" binding:"required"`
	OfferingPlanUpdateable string `json:"offering_plan_updateable" binding:"required"`
	//PlansMin > 0
	PlansMin int `json:"plans_min" binding:"required"`
	//PlansMax >= PlansMin
	PlansMax                   int    `json:"plans_max" binding:"required"`
	PlanMetadata               string `json:"plan_metadata" binding:"required"`
	Free                       string `json:"free" binding:"required"`
	PlanBindable               string `json:"plan_bindable" binding:"required"`
	BindingRotatable           string `json:"binding_rotatable" binding:"required"`
	PlanUpdateable             string `json:"plan_updateable" binding:"required"`
	Schemas                    string `json:"schemas" binding:"required"`
	ServiceInstanceSchema      string `json:"service_instance_schema" binding:"required"`
	ServiceBindingSchema       string `json:"service_binding_schema" binding:"required"`
	MaxPollingDurationMin      int    `json:"max_polling_duration_min" binding:"required"`
	MaxPollingDurationMax      int    `json:"max_polling_duration_max" binding:"required"`
	MaintenanceInfo            string `json:"maintenance_info" binding:"required"`
	MaintenanceInfoVersion     string `json:"maintenance_info_version" binding:"required"`
	MaintenanceInfoDescription string `json:"maintenance_info_description" binding:"required"`
}

//Generates Catalog from file
func GenerateCatalog() (catalog *model.Catalog, err error) {
	var catalogSettings catalogSettings
	catalogSettingsJson, err := os.Open("settings/catalogSettings.json")
	if err != nil {
		log.Println("Error while opening settings/catalogSettings.json! error: " + err.Error())
	} else {
		byteVal, err := ioutil.ReadAll(catalogSettingsJson)
		if err != nil {
			log.Println("Error reading from catalogSettings.json file! error: " + err.Error())
		} else {
			err = json.Unmarshal(byteVal, &catalogSettings)
			if err != nil {
				log.Println("Error unmarshalling the catalogSettings stream to the catalog struct! error: " +
					err.Error())
			}
		}
	}

	return catalog, err
}
