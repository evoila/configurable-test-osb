package tests

import (
	"encoding/json"
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestCatalog(t *testing.T) {
	var catalog interface{}
	catalogJson, err := os.Open("catalog.json")
	if err != nil {
		log.Println("Error while opening catalog file! error: " + err.Error())
		t.Error(err.Error())
	} else {
		byteVal, err := ioutil.ReadAll(catalogJson)
		if err != nil {
			log.Println("Error reading from catalog file! error: " + err.Error())
			t.Error(err.Error())
		} else {
			err = json.Unmarshal(byteVal, &catalog)
			if err != nil {
				log.Println("Error unmarshalling the catalog file to the catalog struct! error: " + err.Error())
				t.Error(err.Error())
			} else {
				brokerController := controller.New()
				var brokerCatalog model.Catalog
				brokerCatalog = *brokerController.ReturnCatalog()
				marshalledJson, err := json.Marshal(brokerCatalog)
				if err != nil {
					t.Error(err.Error())
				} else {
					var unmarshalledAgain interface{}
					err = json.Unmarshal(marshalledJson, &unmarshalledAgain)
					if err != nil {
						t.Error(err.Error())
					} else {
						same := reflect.DeepEqual(unmarshalledAgain, catalog)
						fmt.Printf("Interfaces same? %v\n", same)
					}
				}

			}
		}
	}
	//s1, _ := json.MarshalIndent(catalog, "", "\t")
	//fmt.Print(string(s1))
	//fmt.Println("HEEEEEEEEYYYYY ======================================")

	//s, _ = json.MarshalIndent(brokerCatalog, "", "\t")
	//fmt.Print(string(s))
	//fmt.Println("ANOTHER ONE =====================================")

	//s2, _ := json.MarshalIndent(unmarshalledAgain, "", "\t")
	//fmt.Print(string(s2))

	//fmt.Printf("Is the same? %v\n", strings.Compare(strings.TrimRight(string(s1), "\n"), strings.TrimRight(string(s2), "\n")))//, string(s1) == string(s2))

	//compareRuneArrays(string(s1), string(s2))
}
