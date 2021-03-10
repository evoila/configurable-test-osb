package service

import (
	"fmt"
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

func (deploymentService *DeploymentService) ProvideService(provisionRequest *model.ProvideServiceInstanceRequest,
	instanceID *string) (int, *model.ProvideUpdateServiceInstanceResponse,
	*model.ServiceBrokerError) {
	if deployment, exists := (*deploymentService.serviceInstances)[*instanceID]; exists == true {
		if deploymentService.settings.ProvisionSettings.StatusCodeOK {
			if cmp.Equal(provisionRequest.Parameters, deployment.Parameters()) &&
				deployment.ServiceID() == provisionRequest.ServiceID && deployment.PlanID() == provisionRequest.PlanID &&
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
			Description: "The given plan_id does not exist for this service_id",
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
	deployment, operationID := model.NewServiceDeployment(*instanceID, provisionRequest, deploymentService.settings)

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
	if deployment.UpdatesRunning() {
		return 422, nil, &model.ServiceBrokerError{
			Error:       "ConcurrencyError",
			Description: "The Service Broker does not support concurrent requests while instance is updating.",
		}
	}

	//do the ids HAVE TO match??? this is not directly specified?!
	//if deploymentService.settings.FetchServiceInstanceSettings.OfferingIDMustMatch && *serviceID != deployment.ServiceID() {
	if serviceID != nil && deploymentService.settings.FetchServiceInstanceSettings.OfferingIDMustMatch && *serviceID != deployment.ServiceID() {
		log.Println("Service id of request: " + *serviceID)
		log.Println("Service id of instance: " + deployment.ServiceID())
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	//if deploymentService.settings.FetchServiceInstanceSettings.PlanIDMustMatch && *planID != deployment.PlanID() {
	if planID != nil && deploymentService.settings.FetchServiceInstanceSettings.PlanIDMustMatch && *planID != deployment.PlanID() {
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

func (deploymentService *DeploymentService) UpdateServiceInstance(updateRequest *model.UpdateServiceInstanceRequest,
	instanceID *string, requestID *string) (int, *model.ProvideUpdateServiceInstanceResponse, *model.ServiceBrokerError) {
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*deploymentService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	if !deployment.DeploymentUsable() {

	}
	if *updateRequest.ServiceId != deployment.ServiceID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "InvalidData",
			Description: "this service instance uses a different service offering",
		}
	}

	/*
		IS THIS RIGHT???
		RIGHT NOW THIS CHECK WILL BE DONE IF CONTEXT != NIL BUT SHOULD IT BE DONE ONLY IF OTHER FIELDS == NIL???
	*/
	serviceOffering, _ := deploymentService.catalog.GetServiceOfferingById(deployment.ServiceID())
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
		if updateRequest.PreviousValues.ServiceId != nil && *updateRequest.PreviousValues.ServiceId != deployment.ServiceID() {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "ServiceIDMatch",
				Description: "this service instance uses a different service offering",
			}
		}
		if updateRequest.PreviousValues.PlanId != nil && *updateRequest.PreviousValues.PlanId != deployment.PlanID() {
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
	operationID, _ := deployment.Update(updateRequest, nil)
	var updateServiceInstanceResponse model.ProvideUpdateServiceInstanceResponse

	updateServiceInstanceResponse.DashboardUrl = deployment.DashboardURL()
	requestSettings, err := model.GetRequestSettings(updateRequest.Parameters)
	if err != nil {
		fmt.Println("there has been an error when binding the request parameters in update in deployment service")
	}
	if requestSettings != nil && *requestSettings.AsyncEndpoint {
		updateServiceInstanceResponse.Operation = operationID
	}
	updateServiceInstanceResponse.Metadata = deployment.Metadata()
	//if requestSet
	//BUILD RESPONSE HERE
	//return async status code here
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
			if deploymentService.settings.PollInstanceOperationSettings.DescriptionInResponse { //bindingService.settings.BindingSettings.ReturnDescriptionLastOperation {
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
		} else {
			return 404, nil, &model.ServiceBrokerError{
				Error:       "NotFound",
				Description: "given instance_id was not found",
			}
		}

	}
	if serviceID != nil && *serviceID != deployment.ServiceID() {
		log.Println("Service id of request: " + *serviceID)
		log.Println("Service id of instance: " + deployment.ServiceID())
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if planID != nil && *planID != deployment.PlanID() {
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
		if deploymentService.settings.ProvisionSettings.ShowOperation && *operation.Async() {
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

	statusCode := 200 //ok
	if operation.InstanceUsable() != nil && !*operation.InstanceUsable() && operation.SupposedToFail() {
		statusCode = 410 //gone
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
	if serviceID != nil && *serviceID != deployment.ServiceID() {
		log.Println("Service id of request: " + *serviceID)
		log.Println("Service id of instance: " + deployment.ServiceID())
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if planID != nil && *planID != deployment.PlanID() {
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
		operationID = deployment.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete,
			requestSettings.FailAtOperation, nil, requestSettings.InstanceUsableAfterFail, &deploymentService.lastOperationOfDeletedInstance, instanceID)
	} else {
		operationID = deployment.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete,
			requestSettings.FailAtOperation, nil, requestSettings.InstanceUsableAfterFail, nil, nil)
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
