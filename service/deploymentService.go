package service

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"log"
)

type DeploymentService struct {
	catalog *model.Catalog
	//pointer to settings?
	serviceInstances *map[string]model.ServiceDeployment
}

func NewDeploymentService(catalog *model.Catalog, serviceInstances *map[string]model.ServiceDeployment) *DeploymentService {
	return &DeploymentService{
		catalog:          catalog,
		serviceInstances: serviceInstances,
	}
}

func (deploymentService *DeploymentService) ProvideService(provisionRequest *model.ProvisionRequest, instanceID *string) (int, *model.ServiceBrokerError) {
	//check: id already exists?
	if _, exists := (*deploymentService.serviceInstances)[*instanceID]; exists == true {
		return 409, &model.ServiceBrokerError{
			Error:       "InstanceIDConflict",
			Description: "The given instance_id is already in use",
		}
	}

	//check Service and Plan ID
	serviceOffering, exists := deploymentService.catalog.GetServiceOfferingById(provisionRequest.ServiceID)
	if !exists {
		return 400, &model.ServiceBrokerError{
			Error:       "ServiceIDConflict",
			Description: "The given service_id does not exist in the catalog",
		}
	}
	servicePlan, exists := serviceOffering.GetPlanByID(provisionRequest.PlanID)
	if !exists {
		return 400, &model.ServiceBrokerError{
			Error:       "PlanIDConflict",
			Description: "The given plan_id does not exist for this service_id",
		}
	}
	//Check

	s, _ := json.MarshalIndent(serviceOffering, "", "\t")
	log.Print(string(s))
	log.Println(servicePlan)
	s, _ = json.MarshalIndent(servicePlan, "", "\t")
	log.Print(string(s))

	if provisionRequest.MaintenanceInfo.Version != nil && servicePlan.MaintenanceInfo.Version != nil {
		if *provisionRequest.MaintenanceInfo.Version != *servicePlan.MaintenanceInfo.Version {
			log.Println(*provisionRequest.MaintenanceInfo.Version)
			log.Println(*servicePlan.MaintenanceInfo.Version)
			return 422, &model.ServiceBrokerError{
				Error:       "MaintenanceInfoConflict",
				Description: model.MaintenanceInfoConflict,
			}
		}
	}
	//wenn alles gut:
	deployment := model.ServiceDeployment{InstanceID: *instanceID}
	(*deploymentService.serviceInstances)[*instanceID] = deployment
	log.Println(*deploymentService.serviceInstances)
	return 200, nil
}
