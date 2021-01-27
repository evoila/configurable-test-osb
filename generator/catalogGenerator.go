package generator

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/validators"
	"log"
	"math/rand"
	"os"
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
	log.Println(catalogSettings)
	catalog, err := generateCatalog(&catalogSettings)

	return catalog, err
}

//func (catalogSettings *CatalogSettings) GenerateCatalog() (*model.Catalog, error) {
func generateCatalog(catalogSettings *model.CatalogSettings) (*model.Catalog, error) {
	var catalog model.Catalog
	//var err error
	for i := 0; i < catalogSettings.Amount; i++ {
		catalog.ServiceOfferings = append(catalog.ServiceOfferings, model.ServiceOffering{
			//MUST BE UNIQUE
			Name:        randomString(5),
			Id:          randomString(8) + "-XXXX-XXXX-XXXX-" + randomString(12),
			Description: randomString(6),
			//REST OF SETTINGS
			Bindable: returnBoolean(catalogSettings.OfferingBindable),
		})
	}
	return &catalog, nil
}

func randomString(n int) string {
	const characters = "abcdefghijklmnopqrstuvxyz0123456789"
	randomCharSequence := make([]byte, n)
	for i := range randomCharSequence {
		randomCharSequence[i] = characters[rand.Int63()%int64(len(characters))]
	}
	return string(randomCharSequence)
}

func returnBoolean(frequency string) bool {
	if frequency == "always" {
		return true
	}
	if frequency == "random" {
		if rand.Intn(2) == 1 {
			return true
		}
	}
	return false
}
