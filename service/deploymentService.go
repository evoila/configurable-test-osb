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

func (deploymentService *DeploymentService) ProvideService(provisionRequest *model.ProvideServiceInstanceRequest,
	instanceID *string, acceptsIncomplete bool) (int, *model.ProvideUpdateServiceInstanceResponse,
	*model.ServiceBrokerError) {
	//check: id already exists?
	if deployment, exists := (*deploymentService.serviceInstances)[*instanceID]; exists == true {
		if deploymentService.settings.ProvisionSettings.StatusCodeOK {
			if cmp.Equal(provisionRequest.Parameters, deployment.Parameters()) {
				response := model.NewProvideServiceInstanceResponse(deployment.DashboardURL(),
					deployment.LastOperationID(), deployment.Metadata(), deploymentService.settings)
				return 200, response, nil
			}
		}
		return 409, nil, &model.ServiceBrokerError{
			Error:       "InstanceIDConflict",
			Description: "The given instance_id is already in use",
		}
	}

	//check Service and Plan ID
	serviceOffering, exists := deploymentService.catalog.GetServiceOfferingById(provisionRequest.ServiceID)
	if !exists {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMissing",
			Description: "The given service_id does not exist in the catalog",
		}
	}
	servicePlan, exists := serviceOffering.GetPlanByID(provisionRequest.PlanID)
	if !exists {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMissing",
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
			return 422, nil, &model.ServiceBrokerError{
				Error:       "MaintenanceInfoConflict",
				Description: model.MaintenanceInfoConflict,
			}
		}
	}
	if deploymentService.settings.ProvisionSettings.Async == true {
		/*
			handled in controller
			if !acceptsIncomplete {
				return 422, nil, &model.ServiceBrokerError{
					Error:       "AsyncRequired",
					Description: "This Broker requires client support for asynchronous service operations.",
				}
			}

		*/
		//pass whole request instead of only parmeters???
		deployment := *model.NewServiceDeployment(*instanceID,
			provisionRequest, deploymentService.settings)
		(*deploymentService.serviceInstances)[*instanceID] = deployment

		response := model.NewProvideServiceInstanceResponse(deployment.DashboardURL(),
			deployment.LastOperationID(), deployment.Metadata(), deploymentService.settings)
		return 202, response, nil
	}
	//wenn alles gut:
	deployment := *model.NewServiceDeployment(*instanceID,
		provisionRequest, deploymentService.settings)
	(*deploymentService.serviceInstances)[*instanceID] = deployment
	log.Println(*deploymentService.serviceInstances)
	response := model.NewProvideServiceInstanceResponse(deployment.DashboardURL(),
		deployment.LastOperationID(), deployment.Metadata(), deploymentService.settings)
	return 201, response, nil
}

func (deploymentService *DeploymentService) FetchServiceInstance(instanceID *string, serviceID *string, planID *string) (int,
	*model.FetchingServiceInstanceResponse, *model.ServiceBrokerError) {
	deployment, exists := (*deploymentService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	offering, _ := deploymentService.catalog.GetServiceOfferingById(deployment.ServiceID())
	if !*offering.InstancesRetrievable {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "InstanceNotRetrievable",
			Description: "Service instances of this offering are not retrievable",
		}
	}
	//if deploymentService.catalog.GetServiceOfferingById(deployment.ServiceID())
	if deployment.State() == "in progress" {
		return 422, nil, &model.ServiceBrokerError{
			Error:       "ConcurrencyError",
			Description: "The Service Broker does not support concurrent requests while instance is updating.",
		}
	}
	if deploymentService.settings.FetchServiceInstanceSettings.OfferingIDMustMatch && *serviceID != deployment.ServiceID() {
		log.Println("Service id of request: " + *serviceID)
		log.Println("Service id of instance: " + deployment.ServiceID())
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if deploymentService.settings.FetchServiceInstanceSettings.PlanIDMustMatch && *planID != deployment.PlanID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMatch",
			Description: "The given plan_id does not match the plan_id of the instance",
		}
	}
	response := model.FetchingServiceInstanceResponse{}
	if deploymentService.settings.FetchServiceInstanceSettings.ShowServiceID {
		response.ServiceId = deployment.ServiceID()
	}
	if deploymentService.settings.FetchServiceInstanceSettings.ShowPlanID {
		response.PlanId = deployment.PlanID()
	}
	if deploymentService.settings.FetchServiceInstanceSettings.ShowDashboardURL {
		response.DashboardUrl = deployment.DashboardURL()
	}
	if deploymentService.settings.FetchServiceInstanceSettings.ShowParameters {
		response.Parameters = deployment.Parameters()
	}
	if deploymentService.settings.FetchServiceInstanceSettings.ShowMaintenanceInfo {
		serviceOffering, _ := deploymentService.catalog.GetServiceOfferingById(deployment.ServiceID())
		servicePlan, _ := serviceOffering.GetPlanByID(deployment.PlanID())
		response.MaintenanceInfo = servicePlan.MaintenanceInfo
	}
	if deploymentService.settings.FetchServiceInstanceSettings.ShowMetadata {
		response.Metadata = deployment.Metadata()
	}
	return 200, &response, nil
}
