package main

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"log"
)

/*
var validFrequency = func(fl validator.FieldLevel) bool {
	frequency, ok := fl.Field().Interface().(string)
	log.Println("frequency value validated!")
	if ok {
		if frequency != "never" && frequency != "random" && frequency != "always"{
			return false
		}
	}
	return true
}
*/

/*type bla struct {
	Sack string `json:"sack,omitempty"`
}*/

func main() {

	// Displaying the results

	//myController := controller.New()
	//myController.GenerateCatalog(nil)
	settings, _ := model.NewCatalogSettings()
	catalog, _ := model.NewCatalog(settings)
	s, _ := json.MarshalIndent(catalog, "", "\t")
	log.Print(string(s))
	//value := false
	//var wau bla
	//s, _ := json.MarshalIndent(wau, "", "\t")
	//log.Print(string(s))
	//myController.logCatalog()
	/*
		catalogSettings, err := generator.GenerateCatalog()
		if err != nil {
			log.Println("Error while generating catalogSettings!")
		} else {

			fmt.Println(json.Marshal(catalogSettings))
		}

	*/
	/*var catalogSettings generator.CatalogSettings
	catalogSettingsJson, err := os.Open("settings/catalogSettings.json")
	if err != nil {
		log.Println("Error while opening settings/catalogSettings.json! error: " + err.Error())
	} else {
		decoder := json.NewDecoder(catalogSettingsJson)
		decoder.Decode(&catalogSettings)
		if catalogSettings.OfferingBindable == "" {
			log.Println("Offering bindable is \"\"")
		}
	}

	//log.Println(catalogSettings.OfferingBindable)
	log.Println("catalogSettings values: ")
	log.Println(catalogSettings)
	log.Println("Now validating setting values...")
	if err = validators.ValidateCatalogSettings(&catalogSettings); err != nil{
		log.Println("Invalid setting values! error: " + err.Error())
	} else {

	}

	*/
}
