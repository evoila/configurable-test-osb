package service

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/google/go-cmp/cmp"
	"log"
)

type DeploymentService struct {
	catalog *model.Catalog
	//pointer to settings?
	serviceInstances               *map[string]*model.ServiceDeployment
	settings                       *model.Settings
	lastOperationOfDeletedInstance map[string]*model.Operation
}

func NewDeploymentService(catalog *model.Catalog, serviceInstances *map[string]*model.ServiceDeployment,
	settings *model.Settings) *DeploymentService {
	return &DeploymentService{
		catalog:                        catalog,
		serviceInstances:               serviceInstances,
		settings:                       settings,
		lastOperationOfDeletedInstance: make(map[string]*model.Operation),
	}
}

//deploymentService.ProvideService first checks if the request is valid. It is checked if the instance_id is used (and if so
//it will check, if the service to deploy and the existing one are identical), if the service_offering exists in the
//catalog and if this service offering has the given plan_id. If the request is valid, a new service will be deployed
//by using model.NewServiceDeployment(*instanceID, provisionRequest, deploymentService.settings)
//Returns an int (http status), the actual response and an error if one occurs
func (deploymentService *DeploymentService) ProvideService(provisionRequest *model.ProvideServiceInstanceRequest,
	instanceID *string) (int, *model.ProvideUpdateServiceInstanceResponse,
	*model.ServiceBrokerError) {
	if deployment, exists := (*deploymentService.serviceInstances)[*instanceID]; exists == true {
		if deploymentService.settings.ProvisionSettings.StatusCodeOKPossible {
			if cmp.Equal(provisionRequest.Parameters, deployment.Parameters()) &&
				*deployment.ServiceID() == provisionRequest.ServiceID && *deployment.PlanID() == provisionRequest.PlanID &&
				*deployment.SpaceID() == provisionRequest.SpaceGUID &&
				*deployment.OrganizationID() == provisionRequest.OrganizationGUID {
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
			Description: "The given plan_id does not exist for this service_id in the catalog",
		}
	}
	if provisionRequest.MaintenanceInfo.Version != nil {
		if servicePlan.MaintenanceInfo.Version == nil {
			return 422, nil, &model.ServiceBrokerError{
				Error:       "MaintenanceInfoConflict",
				Description: model.MaintenanceInfoConflict,
			}
		}
		if *provisionRequest.MaintenanceInfo.Version != *servicePlan.MaintenanceInfo.Version {
			log.Println(*provisionRequest.MaintenanceInfo.Version)
			log.Println(*servicePlan.MaintenanceInfo.Version)
			return 422, nil, &model.ServiceBrokerError{
				Error:       "MaintenanceInfoConflict",
				Description: model.MaintenanceInfoConflict,
			}
		}
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(provisionRequest.Parameters)
	deployment, operationID := model.NewServiceDeployment(*instanceID, provisionRequest, deploymentService.settings,
		deploymentService.catalog)
	(*deploymentService.serviceInstances)[*instanceID] = deployment
	response := model.NewProvideServiceInstanceResponse(deployment.DashboardURL(),
		operationID, deployment.Metadata(), deploymentService.settings)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		/*if *requestSettings.FailAtOperation {
			remove deployment from map or leave it there (so that the platform has to deprovision it)?
		}
		*/
		return 202, response, nil
	}
	return 201, response, nil
}

//deploymentService.FetchServiceInstance first checks if the request is valid. It is checked if the instance_id is used,
//if it is retrievable, updating and if serviceID and planID match (if given). If the request is valid, information
//about the service instance will be returned.
//Returns an int (http status), the actual response and an error if one occurs
func (deploymentService *DeploymentService) FetchServiceInstance(instanceID *string, serviceID *string, planID *string) (int,
	*model.FetchingServiceInstanceResponse, *model.ServiceBrokerError) {
	deployment, exists := (*deploymentService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	if serviceID != nil && *serviceID != *deployment.ServiceID() {
		log.Println("Service id of request: " + *serviceID)
		log.Println("Service id of instance: " + *deployment.ServiceID())
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if planID != nil && *planID != *deployment.PlanID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMatch",
			Description: "The given plan_id does not match the plan_id of the instance",
		}
	}
	offering, _ := deploymentService.catalog.GetServiceOfferingById(*deployment.ServiceID())
	if offering.InstancesRetrievable != nil && !*offering.InstancesRetrievable {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "InstanceNotRetrievable",
			Description: "Service instances of this offering are not retrievable",
		}
	}
	if deployment.Blocked() {
		return 422, nil, &model.ServiceBrokerError{
			Error:       "ConcurrencyError",
			Description: model.ConcurrencyError,
		}
	}
	/*
		response := model.FetchingServiceInstanceResponse{}
		if deploymentService.settings.FetchServiceInstanceSettings.ReturnServiceID {
			response.ServiceId = deployment.ServiceID()
		}
		if deploymentService.settings.FetchServiceInstanceSettings.ReturnPlanID {
			response.PlanId = deployment.PlanID()
		}
		if deploymentService.settings.FetchServiceInstanceSettings.ReturnDashboardURL {
			response.DashboardUrl = deployment.DashboardURL()
		}
		if deploymentService.settings.FetchServiceInstanceSettings.ReturnParameters {
			response.Parameters = deployment.Parameters()
		}
		if deploymentService.settings.FetchServiceInstanceSettings.ReturnMaintenanceInfo {
			serviceOffering, _ := deploymentService.catalog.GetServiceOfferingById(*deployment.ServiceID())
			servicePlan, _ := serviceOffering.GetPlanByID(*deployment.PlanID())
			response.MaintenanceInfo = servicePlan.MaintenanceInfo
		}
		if deploymentService.settings.FetchServiceInstanceSettings.ReturnMetadata {
			response.Metadata = deployment.Metadata()
		}

	*/

	return 200, deployment.FetchResponse(), nil
}

func (deploymentService *DeploymentService) UpdateServiceInstance(updateRequest *model.UpdateServiceInstanceRequest,
	instanceID *string) (int, *model.ProvideUpdateServiceInstanceResponse, *model.ServiceBrokerError) {
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*deploymentService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	if *updateRequest.ServiceId != *deployment.ServiceID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "InvalidData",
			Description: "this service instance uses a different service offering",
		}
	}

	serviceOffering, _ := deploymentService.catalog.GetServiceOfferingById(*deployment.ServiceID())
	if updateRequest.Context != nil && serviceOffering.AllowContextUpdates != nil && !*serviceOffering.AllowContextUpdates {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "InvalidData",
			Description: "this service offering does not allow context updates",
		}
	}

	if updateRequest.PlanId != nil {
		if _, exists := serviceOffering.GetPlanByID(*updateRequest.PlanId); !exists {
			return 404, nil, &model.ServiceBrokerError{
				Error:       "NotFound",
				Description: "plan_id was not found for given instance_id",
			}
		}
	}
	if updateRequest.PreviousValues != nil {
		//DEPRECATED (BUT STILL REQUIRED)
		if updateRequest.PreviousValues.ServiceId != nil && *updateRequest.PreviousValues.ServiceId != *deployment.ServiceID() {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "ServiceIDMatch",
				Description: "this service instance uses a different service offering",
			}
		}
		if updateRequest.PreviousValues.PlanId != nil && *updateRequest.PreviousValues.PlanId != *deployment.PlanID() {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "PlanIDMatch",
				Description: "this service instance uses a different service plan",
			}
		}
		//DEPRECATED IN FAVOR OF CONTEXT (BUT STILL REQUIRED) CHECK IF THERE IS A PROFILE (WITH ORGANIZATION_ID) FOR CONTEXT???
		if updateRequest.PreviousValues.OrganizationId != nil && *updateRequest.PreviousValues.OrganizationId != *deployment.OrganizationID() {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "OrganizationIDMatch",
				Description: "this service instance uses a different organization_id",
			}
		}
		//DEPRECATED (BUT STILL REQUIRED)
		if updateRequest.PreviousValues.SpaceID != nil && *updateRequest.PreviousValues.SpaceID != *deployment.SpaceID() {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "SpaceIDMatch",
				Description: "this service instance uses a different space_id",
			}
		}
		//NO INFORMATION ABOUT WHAT TO TO WITH PREVIOUS_VALUES.MAINTENANCE_INFO???
	}
	if updateRequest.MaintenanceInfo != nil && updateRequest.MaintenanceInfo.Version != nil {
		servicePlan, _ := serviceOffering.GetPlanByID(*updateRequest.PlanId)
		if servicePlan.MaintenanceInfo.Version == nil {
			return 422, nil, &model.ServiceBrokerError{
				Error:       "MaintenanceInfoConflict",
				Description: model.MaintenanceInfoConflict,
			}
		}
		if *updateRequest.MaintenanceInfo.Version != *servicePlan.MaintenanceInfo.Version {
			log.Println(*updateRequest.MaintenanceInfo.Version)
			log.Println(*servicePlan.MaintenanceInfo.Version)
			return 422, nil, &model.ServiceBrokerError{
				Error:       "MaintenanceInfoConflict",
				Description: model.MaintenanceInfoConflict,
			}
		}

	}
	operationID, _ := deployment.Update(updateRequest)
	var updateServiceInstanceResponse model.ProvideUpdateServiceInstanceResponse

	updateServiceInstanceResponse.DashboardUrl = deployment.DashboardURL()
	requestSettings, err := model.GetRequestSettings(updateRequest.Parameters)
	if err != nil {
		return 500, nil, &model.ServiceBrokerError{
			Error:       "RequestSettingsError",
			Description: "Error while binding request settings from request parameters.",
		}
	}
	if requestSettings != nil && *requestSettings.AsyncEndpoint {
		updateServiceInstanceResponse.Operation = operationID
	}
	updateServiceInstanceResponse.Metadata = deployment.Metadata()
	if requestSettings != nil && *requestSettings.AsyncEndpoint {
		return 202, &updateServiceInstanceResponse, nil
	}
	if requestSettings != nil && *requestSettings.FailAtOperation {
		return 500, nil, &model.ServiceBrokerError{
			Error:            "OperationFail",
			Description:      "Update operation failed",
			InstanceUsable:   requestSettings.InstanceUsableAfterFail,
			UpdateRepeatable: requestSettings.UpdateRepeatableAfterFail,
		}
	}
	return 200, &updateServiceInstanceResponse, nil
}

func (deploymentService *DeploymentService) PollOperationState(instanceID *string, serviceID *string, planID *string,
	operationName *string) (int, *model.InstanceOperationPollResponse, *model.ServiceBrokerError) {
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*deploymentService.serviceInstances)[*instanceID]
	if !exists {
		operation, instanceDeleted := (*deploymentService).lastOperationOfDeletedInstance[*instanceID]
		if instanceDeleted {
			var responseDescription *string
			if deploymentService.settings.PollInstanceOperationSettings.DescriptionInResponse {
				description := "Default description"
				responseDescription = &description
			}
			if *operation.Async() {
				pollResponse := model.InstanceOperationPollResponse{
					State:       *operation.State(),
					Description: responseDescription,
				}
				//ALWAYS RETURN 410 IF ASYNC DELETION OR ONLY IF OPERATION STATE == "succeeded" (OR != "in progress") ????!!!!
				return 410, &pollResponse, nil
			} else {
				/*
					CORRECT BEHAVIOUR ???!!!
					ALWAYS RETURN FAILED BECAUSE THE DELETION WAS NOT CALLED ASYNC???!
					COULD THIS MEAN THAT ADDING A NEW OPERATION WHEN AN ENDPOINT IS CALLED IS NOT NEEDED????!!!
					CHECK!!!
				*/
				pollResponse := model.InstanceOperationPollResponse{
					State:       "failed",
					Description: responseDescription,
				}
				return 200, &pollResponse, nil
			}
		} //removed else
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}

	}
	if serviceID != nil && *serviceID != *deployment.ServiceID() {
		log.Println("Service id of request: " + *serviceID)
		log.Println("Service id of instance: " + *deployment.ServiceID())
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if planID != nil && *planID != *deployment.PlanID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMatch",
			Description: "The given plan_id does not match the plan_id of the instance",
		}
	}
	var operation *model.Operation
	if operationName != nil {
		operation = deployment.GetOperationByName(*operationName)
		if operation == nil {
			return 404, nil, &model.ServiceBrokerError{
				Error:       "OperationID",
				Description: "The given operation does not exist for the service instance",
			}
		}
	} else {
		operation = deployment.GetLastOperation()
		if deploymentService.settings.ProvisionSettings.ReturnOperation && *operation.Async() {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "MissingOperation",
				Description: "The last operation requires an operation value!",
			}
		}
	}
	var responseDescription *string
	if deploymentService.settings.PollInstanceOperationSettings.DescriptionInResponse {
		description := "Default description"
		responseDescription = &description
	}
	pollResponse := model.InstanceOperationPollResponse{
		State:            *operation.State(),
		Description:      responseDescription,
		InstanceUsable:   operation.InstanceUsable(),
		UpdateRepeatable: operation.UpdateRepeatable(),
	}
	statusCode := 200
	if operation.InstanceUsable() != nil && !*operation.InstanceUsable() && operation.SupposedToFail() {
		statusCode = 410
	}
	return statusCode, &pollResponse, nil
}

func (deploymentService *DeploymentService) Delete(deleteRequest *model.DeleteRequest, instanceID *string,
	serviceID *string, planID *string) (int, *string, *model.ServiceBrokerError) {
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(deleteRequest.Parameters)
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*deploymentService.serviceInstances)[*instanceID]
	if !exists {
		return 410, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	if serviceID != nil && *serviceID != *deployment.ServiceID() {
		log.Println("Service id of request: " + *serviceID)
		log.Println("Service id of instance: " + *deployment.ServiceID())
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if planID != nil && *planID != *deployment.PlanID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMatch",
			Description: "The given plan_id does not match the plan_id of the instance",
		}
	}
	if !deploymentService.settings.ProvisionSettings.AllowDeprovisionWithBindings && deployment.AmountOfBindings() > 0 {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "BlockedByBinding",
			Description: "Deprovision failed because deployment has bindings. Please delete those first or change \"allow_deprovision_with_bindings\" to true",
		}
	}
	var operationID *string
	if !*requestSettings.FailAtOperation {
		delete(*deploymentService.serviceInstances, *instanceID)
		operationID = deployment.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, nil, requestSettings.InstanceUsableAfterFail, &deploymentService.lastOperationOfDeletedInstance, instanceID, false)
	} else {
		operationID = deployment.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, nil, requestSettings.InstanceUsableAfterFail, nil, nil, false)
	}
	var response string
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {

		if deploymentService.settings.ProvisionSettings.ReturnOperationIfAsync {
			response = *operationID
		}
		return 202, &response, nil
	}
	response = "{}"
	return 200, &response, nil
}

//BONUS
func (deploymentService *DeploymentService) CurrentServiceInstances() *map[string]*model.ServiceDeployment {
	return deploymentService.serviceInstances
}
