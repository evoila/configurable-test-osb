package generator

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/validators"
	"log"
	"os"
)

/*
This file generates a random catalog (catalog with random services and offerings) from the file catalogSettings.json, so
that it doesn't need to be created by hand.
*/

//Generates Catalog from file
func GenerateCatalog() (*model.CatalogSettings, error) {
	var catalogSettings model.CatalogSettings
	catalogSettingsJson, err := os.Open("settings/catalogSettings.json")
	if err != nil {
		log.Println("Error while opening settings/catalogSettings.json! error: " + err.Error())
		return nil, err
	}
	decoder := json.NewDecoder(catalogSettingsJson)
	if err = decoder.Decode(&catalogSettings); err != nil {
		return nil, err
	}
	if err = validators.ValidateCatalogSettings(&catalogSettings); err != nil {
		return nil, err
	}
	log.Println("Catalog settings validated!")
	log.Println(catalogSettings)
	//catalog, err := generateCatalog(&catalogSettings)
	return &catalogSettings, err
}

//func (catalogSettings *CatalogSettings) GenerateCatalog() (*model.Catalog, error) {
/*func generateCatalog(catalogSettings *model.CatalogSettings) (*model.Catalog, error) {
	var catalog model.Catalog
	//var err error

	return &catalog, nil
}

*/
