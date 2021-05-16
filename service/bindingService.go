package service

import (
	"github.com/evoila/configurable-test-osb/model"
	"github.com/google/go-cmp/cmp"
)

type BindingService struct {
	serviceInstances              *map[string]*model.ServiceDeployment
	bindingInstances              *map[string]*model.ServiceBinding
	lastOperationOfDeletedBinding map[string]*model.Operation
	settings                      *model.Settings
	catalog                       *model.Catalog
	syncWriteChannel              chan int
	syncReadChannel               chan int
	blockingReaders               int
}

func NewBindingService(serviceInstances *map[string]*model.ServiceDeployment,
	bindingInstances *map[string]*model.ServiceBinding, settings *model.Settings, catalog *model.Catalog) *BindingService {
	syncWriteChannel := make(chan int, 1)
	syncWriteChannel <- 1
	syncReadChannel := make(chan int, 1)
	syncReadChannel <- 1
	return &BindingService{
		serviceInstances:              serviceInstances,
		bindingInstances:              bindingInstances,
		lastOperationOfDeletedBinding: make(map[string]*model.Operation),
		settings:                      settings,
		catalog:                       catalog,
		syncWriteChannel:              syncWriteChannel,
		syncReadChannel:               syncReadChannel,
		blockingReaders:               0,
	}
}

//bindingService.CreateBinding first checks if the request is valid. It is checked if the instance_id exists,
//if the instance is bindable, if the bindingID is already in use (and if so it will check, if the binding to create
//and the existing one are identical) and if serviceID and planID match. If the request is valid,
//model.NewServiceBinding(bindingID *string, bindingRequest *CreateBindingRequest, settings *Settings,
//catalog *Catalog, bindingInstances *map[string]*ServiceBinding, deployment *ServiceDeployment)
//will be called, which creates a binding to a deployment.
//Returns an int (http status), the actual response and an error if one occurs
func (bindingService *BindingService) CreateBinding(bindingRequest *model.CreateBindingRequest, instanceID *string,
	bindingID *string) (int, *model.CreateRotateFetchBindingResponse, *model.ServiceBrokerError) {
	deployment, exists := (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	if deployment.IsDeploying() {
		return 422, nil, &model.ServiceBrokerError{
			Error:       "ConcurrencyError",
			Description: model.ConcurrencyError,
		}
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(bindingRequest.Parameters)
	<-bindingService.syncWriteChannel
	defer bindingService.writeToSyncChannel()
	if binding, exists := (*bindingService.bindingInstances)[*bindingID]; exists == true {
		if bindingService.settings.BindingSettings.StatusCodeOKPossible && !*requestSettings.AsyncEndpoint {
			if cmp.Equal(bindingRequest.Parameters, binding.Parameters()) &&
				cmp.Equal(bindingRequest.Context, binding.Context()) &&
				*bindingRequest.ServiceID == *binding.ServiceID() &&
				*bindingRequest.PlanID == *binding.PlanID() &&
				cmp.Equal(bindingRequest.AppGUID, binding.AppGuid()) &&
				cmp.Equal(bindingRequest.BindResource, binding.BindResource()) {
				if bindingService.settings.BindingSettings.ReturnBindingInformationOnce && binding.InformationReturned() {
					return 200, &model.CreateRotateFetchBindingResponse{}, nil
				}
				response := binding.Response()
				response.Parameters = nil
				return 200, response, nil
			}
		}
		return 409, nil, &model.ServiceBrokerError{
			Error:       "InstanceIDConflict",
			Description: "The given binding_id is already in use",
		}
	}
	if *bindingRequest.ServiceID != *deployment.ServiceID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if *bindingRequest.PlanID != *deployment.PlanID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMatch",
			Description: "The given plan_id does not match the plan_id of the instance",
		}
	}
	offering, _ := bindingService.catalog.GetServiceOfferingById(*bindingRequest.ServiceID)
	plan, _ := offering.GetPlanByID(*bindingRequest.PlanID)
	if *offering.Bindable == false || plan.Bindable != nil && *plan.Bindable == false {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "InstanceNotBindable",
			Description: "Service instances of this offering are not bindable",
		}
	}
	binding, operationID := model.NewServiceBinding(bindingID, bindingRequest, bindingService.settings, bindingService.catalog, bindingService.bindingInstances, deployment)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		var response model.CreateRotateFetchBindingResponse
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync {
			response.Operation = operationID
		}
		return 202, &response, nil
	}
	binding.SetInformationReturned(true)
	response := binding.Response()
	response.Parameters = nil
	return 201, binding.Response(), nil
}

//bindingService.RotateBinding first checks if the request is valid. It is checked if the instance_id exists,
//if the instance is bindable and if the bindingID is already in use. If the request is valid,
//ServiceBinding.RotateBinding(rotateBindingRequest *RotateBindingRequest, deployment *ServiceDeployment)
//will be called, which creates a binding to a deployment with the values from an old binding.
//Returns an int (http status), the actual response and an error if one occurs
func (bindingService *BindingService) RotateBinding(rotateBindingRequest *model.RotateBindingRequest, instanceID *string,
	bindingID *string) (int, *model.CreateRotateFetchBindingResponse, *model.ServiceBrokerError) {
	deployment, exists := (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "Given instance_id was not found",
		}
	}
	oldBinding, exists := deployment.GetBinding(rotateBindingRequest.PredecessorBindingId)
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "Given predecessor_binding_id was not found",
		}
	}
	bindingService.beginRead()
	existingBinding, exists := (*bindingService.bindingInstances)[*bindingID]
	bindingService.endRead()
	if exists {
		var requestSettings *model.RequestSettings
		requestSettings, _ = model.GetRequestSettings(rotateBindingRequest.Parameters)
		if bindingService.settings.BindingSettings.StatusCodeOKPossible && !*requestSettings.AsyncEndpoint {
			if cmp.Equal(oldBinding.Parameters(), existingBinding.Parameters()) &&
				cmp.Equal(oldBinding.Context(), existingBinding.Context()) &&
				*oldBinding.ServiceID() == *existingBinding.ServiceID() &&
				*oldBinding.PlanID() == *existingBinding.PlanID() &&
				cmp.Equal(oldBinding.AppGuid(), existingBinding.AppGuid()) &&
				cmp.Equal(oldBinding.BindResource(), existingBinding.BindResource()) {
				if bindingService.settings.BindingSettings.ReturnBindingInformationOnce && existingBinding.InformationReturned() {
					return 200, &model.CreateRotateFetchBindingResponse{}, nil
				}
				response := existingBinding.Response()
				response.Parameters = nil
				return 200, response, nil
			}
		}
		return 409, nil, &model.ServiceBrokerError{
			Error:       "InstanceIDConflict",
			Description: "The given binding_id is already in use",
		}
	}

	newBinding, operationID := oldBinding.RotateBinding(rotateBindingRequest, deployment, bindingID)
	(*bindingService.bindingInstances)[*bindingID] = newBinding
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(rotateBindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		var response model.CreateRotateFetchBindingResponse
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync {
			response.Operation = operationID
		}
		return 202, &response, nil
	}
	newBinding.SetInformationReturned(true)
	response := newBinding.Response()
	response.Parameters = nil
	return 201, newBinding.Response(), nil
}

//bindingService.FetchBinding first checks if the request is valid. It is checked if the instance_id and binding_id exist,
//and if service and plan id match with the ones of the binding, if given. If the request is valid and the binding can be fetched,
//binding information will be returned.
//Returns an int (http status), the actual response and an error if one occurs
func (bindingService *BindingService) FetchBinding(instanceID *string, bindingID *string, serviceID *string,
	planID *string) (int, *model.CreateRotateFetchBindingResponse, *model.ServiceBrokerError) {
	deployment, exists := (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "Given instance_id was not found",
		}
	}
	bindingService.beginRead()
	binding, exists := deployment.GetBinding(bindingID)
	bindingService.endRead()
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "This service instance does not use a binding with the given binding_id",
		}
	}
	if serviceID != nil && *serviceID != "" {
		if *deployment.ServiceID() != *serviceID {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "ServiceIDMatch",
				Description: "The given service_id does not match the service_id of the instance",
			}
		}
	}
	if planID != nil && *planID != "" {
		if *deployment.PlanID() != *planID {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "ServiceIDMatch",
				Description: "The given plan_id does not match the plan_id of the instance",
			}
		}
	}
	offering, _ := bindingService.catalog.GetServiceOfferingById(*deployment.ServiceID())
	if offering.BindingsRetrievable == nil || offering.BindingsRetrievable != nil && !*offering.BindingsRetrievable {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "BindingNotRetrievable",
			Description: "Service bindings of this offering are not retrievable",
		}
	}
	if bindingService.settings.BindingSettings.ReturnBindingInformationOnce && binding.InformationReturned() {
		return 200, &model.CreateRotateFetchBindingResponse{}, nil
	}
	response := binding.Response()
	if bindingService.settings.BindingSettings.ReturnParameters {
		response.Parameters = binding.Parameters()
	}
	binding.SetInformationReturned(true)
	return 200, response, nil
}

//bindingService.PollOperationState first checks if the request is valid. It is checked if the instance_id and binding_id exist,
//and if service and plan id match with the ones of the service instance, if given.
//If an operationName is given, the function will try to look up the operation with the operationName as key.
//If the setting to return the operation if async (like when provisioning async) is true and the last operation was
//async, the correct operationName is REQUIRED in order to return the last operation.
//If the request is valid and the binding exists, the state of the operation will be returned.
//If deleted, the last operation can still be accessed through deploymentService.lastOperationOfDeletedInstance to get
//information about the deletion process.
//Returns an int (http status), the actual response and an error if one occurs
func (bindingService *BindingService) PollOperationState(instanceID *string, bindingID *string, serviceID *string,
	planID *string, operationName *string) (int, *model.InstanceOperationPollResponse, *model.ServiceBrokerError) {
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "Given instance_id was not found",
		}
	}
	if serviceID != nil && *serviceID != *deployment.ServiceID() {
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
	bindingService.beginRead()
	_, bindingExists := (*bindingService.bindingInstances)[*bindingID]
	bindingService.endRead()
	if !bindingExists {
		operation, bindingDeleted := (*bindingService).lastOperationOfDeletedBinding[*bindingID]
		if bindingDeleted {
			if !deployment.BindingDeleted(bindingID) {
				return 404, nil, &model.ServiceBrokerError{
					Error:       "NotFound",
					Description: "Given binding_id was not found for this instance_id",
				}
			}
			var responseDescription *string
			if bindingService.settings.BindingSettings.ReturnDescriptionLastOperation {
				description := "Default description"
				responseDescription = &description
			}
			if *operation.Async() {
				pollResponse := model.InstanceOperationPollResponse{
					State:       *operation.State(),
					Description: responseDescription,
				}
				return 410, &pollResponse, nil
			} else {
				pollResponse := model.InstanceOperationPollResponse{
					State:       "failed",
					Description: responseDescription,
				}
				return 200, &pollResponse, nil
			}
		}
	}
	binding, exists := deployment.GetBinding(bindingID)
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "Given binding_id was not found for this instance_id",
		}
	}
	var operation *model.Operation
	if operationName != nil {
		operation = binding.GetOperationByName(*operationName)
		if operation == nil {
			return 404, nil, &model.ServiceBrokerError{
				Error:       "OperationID",
				Description: "The given operation does not exist for the service instance",
			}
		}
	} else {
		operation = binding.GetLastOperation()
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync && *operation.Async() {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "MissingOperation",
				Description: "The last operation requires an operation value!",
			}
		}
	}
	var responseDescription *string
	if bindingService.settings.BindingSettings.ReturnDescriptionLastOperation {
		description := "Default description"
		responseDescription = &description
	}
	pollResponse := model.InstanceOperationPollResponse{
		State:       *operation.State(),
		Description: responseDescription,
	}
	statusCode := 200
	if operation.InstanceUsable() != nil && !*operation.InstanceUsable() && operation.SupposedToFail() {
		statusCode = 410
	}
	return statusCode, &pollResponse, nil
}

//bindingService.Unbind first checks if the request is valid. It is checked if the instance_id and binding_id exist,
//and if service and plan id match with the ones of the binding If the request is valid the binding will be
//deleted from the map and from the deployment by calling serviceDeployment.RemoveBinding(bindingID *string)
//Returns an int (http status), the actual response and an error if one occurs
func (bindingService *BindingService) Unbind(deleteRequest *model.DeleteRequest, instanceID *string, bindingID *string,
	serviceID *string, planID *string) (int, *model.OperationResponse, *model.ServiceBrokerError) {
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(deleteRequest.Parameters)
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 410, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "Given instance_id was not found",
		}
	}
	<-bindingService.syncWriteChannel
	defer bindingService.writeToSyncChannel()
	binding, exists := deployment.GetBinding(bindingID)
	if !exists {
		return 410, nil, &model.ServiceBrokerError{
			Error:       "Gone",
			Description: "Given binding_id does not exist",
		}
	}
	if serviceID != nil && *serviceID != *deployment.ServiceID() {
		return 412, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if planID != nil && *planID != *deployment.PlanID() {
		return 412, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMatch",
			Description: "The given plan_id does not match the plan_id of the instance",
		}
	}

	var operationID *string
	if !*requestSettings.FailAtOperation {
		deployment.RemoveBinding(bindingID)
		delete(*bindingService.bindingInstances, *bindingID)
		operationID = binding.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, &bindingService.lastOperationOfDeletedBinding, bindingID)
	} else {
		operationID = binding.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, nil, nil)
	}

	var operationResponse model.OperationResponse
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync {
			operationResponse.Operation = operationID
		}
		return 202, &operationResponse, nil
	}
	return 200, &operationResponse, nil
}

func (bindingService *BindingService) CurrentBindings() *map[string]*model.ServiceBinding {
	return bindingService.bindingInstances
}

func (bindingService *BindingService) writeToSyncChannel() {
	bindingService.syncWriteChannel <- 1
}

func (bindingService *BindingService) beginRead() {
	<-bindingService.syncReadChannel
	bindingService.blockingReaders++
	if bindingService.blockingReaders == 1 {
		<-bindingService.syncWriteChannel
	}
	bindingService.syncReadChannel <- 1
}

func (bindingService *BindingService) endRead() {
	<-bindingService.syncReadChannel
	bindingService.blockingReaders--
	if bindingService.blockingReaders == 0 {
		bindingService.syncWriteChannel <- 1
	}
	bindingService.syncReadChannel <- 1
}
