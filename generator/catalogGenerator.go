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
		catalog.ServiceOfferings = append(catalog.ServiceOfferings, model.ServiceOffering{
			//MUST BE UNIQUE
			Name:        randomString(5),
			Id:          randomString(8) + "-XXXX-XXXX-XXXX-" + randomString(12),
			Description: randomString(6),
			//REST OF SETTINGS
			Tags: selectRandomTags(tags, catalogSettings.TagsMin, catalogSettings.TagsMax),
			//Requires:
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

/*func randomRequires(min int, max int)  {
	requires := [3]string{"syslog_drain", "route_forwarding", "volume_mount"}
}*/
