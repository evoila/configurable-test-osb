package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	_ "github.com/MaxFuhrich/serviceBrokerDummy/validators"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

/*
TO DO
-Securing via TLS?
	-Spec: "This specification does not specify how Platform and Service Brokers agree on other
	methods of authentication"
*/

func Run() {

	//Remove in future when other controllers exist?
	testController := controller.New()
	catalog, err := makeCatalog()
	if err != nil {
		log.Println("There has been an error while creating the catalog!", err.Error())
	} else {
		settings, err := makeSettings()
		if err != nil {
			log.Println("There has been an error while creating the settings!", err.Error())
		} else {
			log.Println()
			catalogService := service.NewCatalogService(catalog)
			catalogController := controller.NewCatalogController(&catalogService, settings)
			//PUT THIS TO A DIFFERENT PLACE
			//Default router, should be changed?
			r := gin.Default()

			//ENDPOINTS HERE

			//Test
			r.GET("/", testController.Hello)
			r.POST("/", testController.TestBind)
			//new endpoints with new controllers
			r.GET("/v2/catalog", catalogController.GetCatalog)

			//Replace when new controllers are implemented?
			//Catalog
			//r.GET("/v2/catalog", testController.GetCatalog)
			/*
				//Polling last operation for service instances
				r.GET("/v2/service_instances/:instance_id/last_operation", testController.LastOpServiceInstance)

				//Polling last operation for service binding
				r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", testController.LastOpServiceBinding)

				//Provisioning (of service)
				r.PUT("/v2/service_instances/:instance_id", testController.ProvideService)

				//Fetching service instance
				r.GET("/v2/service_instances/:instance_id", testController.FetchServiceInstance)

				//Updating service instance
				r.PATCH("/v2/service_instances/:instance_id", testController.UpdateServiceInstance)

				//Request (creating service binding)
				r.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", testController.CreateServiceBinding)

				//Request (rotating service binding)
				r.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", testController.RotateServiceBinding)

				//Fetching service binding
				r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", testController.FetchServiceBinding)

				//Unbinding
				r.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", testController.Unbind)

				//Deprovisioning
				r.DELETE("/v2/service_instances/:instance_id", testController.Deprovide)
			*/
			//Generating new random catalog from catalogSettings.json
			//r.GET("/generate_catalog", testController.GenerateCatalog)
			err = r.Run()
			if err != nil {
				log.Println("Error: " + err.Error())
				fmt.Println("Error: " + err.Error())
			}
		}

	}
}

func makeCatalog() (*model.Catalog, error) {
	var catalog model.Catalog
	//FILEPATH
	catalogJson, err := os.Open("config/catalog.json")
	if err != nil {
		return nil, errors.New("error while opening catalog file! error: " + err.Error())
	}
	byteVal, err := ioutil.ReadAll(catalogJson)
	if err != nil {
		return nil, errors.New("error reading from catalog file! error: " + err.Error())
	}
	err = json.Unmarshal(byteVal, &catalog)
	if err != nil {
		return nil, errors.New("error unmarshalling the catalog file to the catalog struct! error: " + err.Error())
	}

	return &catalog, nil
}

func makeSettings() (*model.Settings, error) {
	var settings model.Settings
	//FILEPATH
	brokerSettingsJson, err := os.Open("config/brokerSettings.json")
	if err != nil {
		return nil, errors.New("error while opening settings file! error: " + err.Error())
	}
	byteVal, err := ioutil.ReadAll(brokerSettingsJson)
	if err != nil {
		return nil, errors.New("error reading from settings file! error: " + err.Error())
	}
	err = json.Unmarshal(byteVal, &settings)
	if err != nil {
		return nil, errors.New("error unmarshalling the settings file to the catalog struct! error: " + err.Error())
	}

	version := strings.Split(settings.HeaderSettings.BrokerVersion, ".")
	if len(version) != 2 {
		return nil, errors.New("the format of the broker version must be \"MAJOR.MINOR\"")
	}
	if _, err = strconv.Atoi(version[0]); err != nil {
		return nil, errors.New("version \"MAJOR\" in \"MAJOR.MINOR\" must be an integer")
	}
	if _, err = strconv.Atoi(version[1]); err != nil {
		return nil, errors.New("version \"MINOR\" in \"MAJOR.MINOR\" must be an integer")
	}
	return &settings, nil
}
