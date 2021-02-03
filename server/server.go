package server

import (
	"encoding/json"
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	_ "github.com/MaxFuhrich/serviceBrokerDummy/validators"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
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
	catalog := makeCatalog()
	catalogService := service.NewCatalogService(catalog)
	catalogController := controller.NewCatalogController(&catalogService)
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
	err := r.Run()
	if err != nil {
		log.Println("Error: " + err.Error())
		fmt.Println("Error: " + err.Error())
	}
}

func makeCatalog() *model.Catalog {
	var catalog model.Catalog
	//FILEPATH
	catalogJson, err := os.Open("catalog/catalog.json")
	if err != nil {
		log.Println("Error while opening catalog file! error: " + err.Error())
	} else {
		byteVal, err := ioutil.ReadAll(catalogJson)
		if err != nil {
			log.Println("Error reading from catalog file! error: " + err.Error())
		} else {
			err = json.Unmarshal(byteVal, &catalog)
			if err != nil {
				log.Println("Error unmarshalling the catalog file to the catalog struct! error: " + err.Error())
			}
		}
	}
	return &catalog
}
