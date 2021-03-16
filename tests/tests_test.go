package tests

import (
	"encoding/json"
	"fmt"
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
				var brokerCatalog model.Catalog
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
}
