package generator

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/validators"
	"log"
	"math/rand"
	"os"
	"time"
)

/*
This file generates a random catalog (catalog with random services and offerings) from the file catalogSettings.json, so
that it doesn't need to be created by hand.
*/

//Generates Catalog from file
func GenerateCatalog() (*model.Catalog, error) {
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

	s, _ := json.MarshalIndent(catalogSettings, "", "\t")
	log.Print(string(s))

	catalog, err := generateCatalog(&catalogSettings)
	//log.Println(catalog)

	return catalog, err
}

//func (catalogSettings *CatalogSettings) GenerateCatalog() (*model.Catalog, error) {
func generateCatalog(catalogSettings *model.CatalogSettings) (*model.Catalog, error) {
	var catalog model.Catalog
	//var err error
	//create tags
	rand.Seed(time.Now().UnixNano())
	var tags []string
	for i := 0; i < catalogSettings.TagsMax; i++ {
		tag := randomString(4)
		for containsString(tags, tag) {
			tag = randomString(4)
		}
		tags = append(tags, tag)
		//append(tags, randomString(4))
	}
	for i := 0; i < catalogSettings.Amount; i++ {
		catalog.ServiceOfferings = append(catalog.ServiceOfferings, *returnServiceOffering(catalogSettings, tags))
	}
	return &catalog, nil
}

//func returnServicePlans(catalogSettins *model.CatalogSettings) []model.ServicePlan
