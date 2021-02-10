package service

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/google/go-cmp/cmp"
	"log"
)

type DeploymentService struct {
	catalog *model.Catalog
	//pointer to settings?
	serviceInstances *map[string]model.ServiceDeployment
	settings         *model.Settings
}

func NewDeploymentService(catalog *model.Catalog, serviceInstances *map[string]model.ServiceDeployment,
	settings *model.Settings) *DeploymentService {
	return &DeploymentService{
		catalog:          catalog,
		serviceInstances: serviceInstances,
		settings:         settings,
	}
}

func (deploymentService *DeploymentService) ProvideService(provisionRequest *model.ProvisionRequest,
	instanceID *string, acceptsIncomplete bool) (int, *model.ServiceBrokerError) {
	//check: id already exists?
	if deployment, exists := (*deploymentService.serviceInstances)[*instanceID]; exists == true {
		if deploymentService.settings.ProvisionSettings.StatusCodeOK {
			if cmp.Equal(provisionRequest.Parameters, deployment.Parameters()) {
				return 200, nil
			}
		}
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

	//s, _ := json.MarshalIndent(serviceOffering, "", "\t")
	//log.Print(string(s))
	//log.Println(servicePlan)
	//s, _ = json.MarshalIndent(servicePlan, "", "\t")
	//log.Print(string(s))

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
	if deploymentService.settings.GeneralSettings.Async == true {
		if !acceptsIncomplete {
			return 422, &model.ServiceBrokerError{
				Error:       "AsyncRequired",
				Description: "This Broker requires client support for asynchronous service operations.",
			}
		}
		//statt wert von async direkt settings mitgeben??? dann aber weniger flexibel?
		//deployment :=
		(*deploymentService.serviceInstances)[*instanceID] = *model.NewServiceDeployment(*instanceID,
			provisionRequest.Parameters, deploymentService.settings.GeneralSettings.Async, 0)
		return 202, nil
	}
	//wenn alles gut:
	(*deploymentService.serviceInstances)[*instanceID] = *model.NewServiceDeployment(*instanceID,
		provisionRequest.Parameters, deploymentService.settings.GeneralSettings.Async, 0)
	log.Println(*deploymentService.serviceInstances)
	return 201, nil
}
