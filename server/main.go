package main

import (
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	"github.com/gin-gonic/gin"
	"log"
)

/*
TO DO
-Securing via TLS?
	-Spec: "This specification does not specify how Platform and Service Brokers agree on other
	methods of authentication"
 */

func main() {

	//PUT THIS TO A DIFFERENT PLACE
	//Default router, should be changed?
	r := gin.Default()

	//ENDPOINTS HERE

	//Test
	r.GET("/", controller.Hello)

	//Catalog
	r.GET("/v2/catalog", controller.GetCatalog)

	//Polling last operation for service instances
	r.GET("/v2/service_instances/:instance_id/last_operation", controller.LastOpServiceInstance)

	//Polling last operation for service binding
	r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", controller.LastOpServiceBinding)

	//Provisioning (of service)
	r.PUT("/v2/service_instances/:instance_id", controller.ProvideService)

	//Fetching service instance
	r.GET("/v2/service_instances/:instance_id", controller.FetchServiceInstance)

	//Updating service instance
	r.PATCH("/v2/service_instances/:instance_id", controller.UpdateServiceInstance)

	//Request (creating service binding)
	r.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", controller.CreateServiceBinding)

	//Request (rotating service binding)
	r.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", controller.RotateServiceBinding)

	//Fetching service binding
	r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", controller.FetchServiceBinding)

	//Unbinding
	r.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", controller.Unbind)

	//Deprovisioning
	r.DELETE("/v2/service_instances/:instance_id", controller.Deprovide)

	err := r.Run()
	if err != nil {
		log.Println("Error: " + err.Error())
		fmt.Println("Error: " + err.Error())
	}
}
