package server

import (
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	_ "github.com/MaxFuhrich/serviceBrokerDummy/validators"
	"github.com/gin-gonic/gin"
	"log"
)

/*
TO DO
-Securing via TLS?
	-Spec: "This specification does not specify how Platform and Service Brokers agree on other
	methods of authentication"
*/

func Run() {
	brokerController := controller.New()
	//PUT THIS TO A DIFFERENT PLACE
	//Default router, should be changed?
	r := gin.Default()

	//ENDPOINTS HERE

	//Test
	r.GET("/", brokerController.Hello)
	r.POST("/", brokerController.TestBind)
	/*
		//Catalog
		r.GET("/v2/catalog", brokerController.GetCatalog)

		//Polling last operation for service instances
		r.GET("/v2/service_instances/:instance_id/last_operation", brokerController.LastOpServiceInstance)

		//Polling last operation for service binding
		r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", brokerController.LastOpServiceBinding)

		//Provisioning (of service)
		r.PUT("/v2/service_instances/:instance_id", brokerController.ProvideService)

		//Fetching service instance
		r.GET("/v2/service_instances/:instance_id", brokerController.FetchServiceInstance)

		//Updating service instance
		r.PATCH("/v2/service_instances/:instance_id", brokerController.UpdateServiceInstance)

		//Request (creating service binding)
		r.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", brokerController.CreateServiceBinding)

		//Request (rotating service binding)
		r.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", brokerController.RotateServiceBinding)

		//Fetching service binding
		r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", brokerController.FetchServiceBinding)

		//Unbinding
		r.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", brokerController.Unbind)

		//Deprovisioning
		r.DELETE("/v2/service_instances/:instance_id", brokerController.Deprovide)
	*/
	//Generating new random catalog from catalogSettings.json
	r.GET("/generate_catalog", brokerController.GenerateCatalog)
	err := r.Run()
	if err != nil {
		log.Println("Error: " + err.Error())
		fmt.Println("Error: " + err.Error())
	}
}
