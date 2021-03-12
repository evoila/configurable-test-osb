package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func Run() {
	catalog, err := makeCatalog()
	if err != nil {
		log.Println("There has been an error while creating the catalog!", err.Error())
	} else {
		settings, err := makeSettings()
		if err != nil {
			log.Println("There has been an error while creating the settings!", err.Error())
		} else {
			var serviceInstances map[string]*model.ServiceDeployment
			serviceInstances = make(map[string]*model.ServiceDeployment)
			catalogService := service.NewCatalogService(catalog)
			catalogController := controller.NewCatalogController(&catalogService, settings)
			deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
			deploymentController := controller.NewDeploymentController(deploymentService, settings)

			var bindingInstances map[string]*model.ServiceBinding
			bindingInstances = make(map[string]*model.ServiceBinding)
			bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
			bindingController := controller.NewBindingController(bindingService, settings)

			middleware := controller.NewMiddleware(settings)
			//PUT THIS TO A DIFFERENT PLACE
			//Default router, should be changed?
			r := gin.Default()
			//check if requestid is ALWAYS required and therefore the first check can be omitted???!!!!
			if settings.HeaderSettings.RequestIDRequired && settings.HeaderSettings.LogRequestID {
				r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
					return fmt.Sprintf("%s\n",
						param.Request.Header.Values("X-Broker-API-Request-Identity"),
					)
				}))
			}
			r.Use(middleware.BindAndCheckHeader)

			//BONUS
			r.GET("/v2/catalog/generate", catalogController.GenerateCatalog)
			r.GET("/v2/service_instances", deploymentController.CurrentServiceInstances)
			r.GET("/v2/service_bindings", bindingController.CurrentBindings)

			//catalog
			r.GET("/v2/catalog", catalogController.GetCatalog)

			//Provisioning (of service)
			r.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)

			//fetching service instances
			r.GET("/v2/service_instances/:instance_id", deploymentController.FetchServiceInstance)

			//Updating service instance
			r.PATCH("/v2/service_instances/:instance_id", deploymentController.UpdateServiceInstance)

			//Polling last operation for service instances
			r.GET("/v2/service_instances/:instance_id/last_operation", deploymentController.PollOperationState)

			//Creating service binding (and rotate, since it uses the same endpoint
			r.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)

			//Fetching service binding
			r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)

			//Polling last operation for service binding
			r.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", bindingController.PollOperationState)

			//Unbinding
			r.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.Unbind)

			//Deprovisioning
			r.DELETE("/v2/service_instances/:instance_id", deploymentController.Delete)

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
	//FILEPATH???!!!
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
	if settings.BindingSettings.BindingEndpointSettings.ProtocolValue != "tcp" &&
		settings.BindingSettings.BindingEndpointSettings.ProtocolValue != "udp" &&
		settings.BindingSettings.BindingEndpointSettings.ProtocolValue != "all" {
		return nil, errors.New("protocol_value must be either \"tct\", \"udp\", or \"all\"")
	}
	return &settings, nil
}
