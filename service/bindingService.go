package service

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/google/go-cmp/cmp"
	"log"
)

type BindingService struct {
	//POINTER TO SERVICE INSTANCES/DEPLOYMENTS???? PRETTY SURE YES
	serviceInstances              *map[string]*model.ServiceDeployment
	bindingInstances              *map[string]*model.ServiceBinding
	lastOperationOfDeletedBinding map[string]*model.Operation
	settings                      *model.Settings
	catalog                       *model.Catalog
}

func NewBindingService(serviceInstances *map[string]*model.ServiceDeployment,
	bindingInstances *map[string]*model.ServiceBinding, settings *model.Settings, catalog *model.Catalog) *BindingService {
	return &BindingService{
		serviceInstances:              serviceInstances,
		bindingInstances:              bindingInstances,
		lastOperationOfDeletedBinding: make(map[string]*model.Operation),
		settings:                      settings,
		catalog:                       catalog,
	}
}

//bindingService.CreateBinding first checks if the request is valid. It is checked if the instance_id exists,
//if the instance is bindable, if the bindingID is already in use (and if so it will check, if the service to deploy
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
	if deployment.Blocked() {
		return 422, nil, &model.ServiceBrokerError{
			Error:       "ConcurrencyError",
			Description: model.ConcurrencyError,
		}
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(bindingRequest.Parameters)
	if binding, exists := (*bindingService.bindingInstances)[*bindingID]; exists == true {
		if bindingService.settings.BindingSettings.StatusCodeOKPossible && !*requestSettings.AsyncEndpoint {
			log.Println("binding with given id exists. now comparing equality of requested binding and existing binding")
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

func (bindingService *BindingService) RotateBinding(rotateBindingRequest *model.RotateBindingRequest, instanceID *string,
	bindingID *string) (int, *model.CreateRotateFetchBindingResponse, *model.ServiceBrokerError) {
	deployment, exists := (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	if _, exists := (*bindingService.bindingInstances)[*bindingID]; exists == true {
		return 409, nil, &model.ServiceBrokerError{
			Error:       "InstanceIDConflict",
			Description: "The given binding_id is already in use",
		}
	}
	oldBinding, exists := (*bindingService.bindingInstances)[*rotateBindingRequest.PredecessorBindingId]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given predecessor_binding_id was not found",
		}
	}
	newBinding, operationID := oldBinding.RotateBinding(rotateBindingRequest, deployment)
	(*bindingService.bindingInstances)[*bindingID] = newBinding
	//deployment.AddBinding(newBinding)
	//response.ReturnOperationIfAsync = "hello"
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(rotateBindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		var response model.CreateRotateFetchBindingResponse
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync {
			response.Operation = operationID
			//this is still ok, the response in the binding is only used when not async created or when fetched
		}
		return 202, &response, nil
	}
	//binding, operationID  := model.NewServiceBinding(*bindingID, rotateBindingRequest, bindingService.settings, bindingService.catalog)
	newBinding.SetInformationReturned(true)
	response := newBinding.Response()
	response.Parameters = nil
	return 201, newBinding.Response(), nil
}

func (bindingService *BindingService) FetchBinding(instanceID *string, bindingID *string, serviceID *string,
	planID *string) (int, *model.CreateRotateFetchBindingResponse, *model.ServiceBrokerError) {
	deployment, exists := (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	binding, exists := deployment.GetBinding(bindingID)
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "this service instance does not use a binding with the given binding_id",
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
	if offering.BindingsRetrievable != nil && *offering.BindingsRetrievable == false {
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

func (bindingService *BindingService) PollOperationState(instanceID *string, bindingID *string, serviceID *string,
	planID *string, operationName *string) (int, *model.InstanceOperationPollResponse, *model.ServiceBrokerError) {
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	if _, bindingExists := (*bindingService.bindingInstances)[*bindingID]; !bindingExists {
		log.Println("binding does not exist, now checking deleted bindings")
		operation, bindingDeleted := (*bindingService).lastOperationOfDeletedBinding[*bindingID]
		if bindingDeleted {
			if !deployment.BindingDeleted(bindingID) {
				return 404, nil, &model.ServiceBrokerError{
					Error:       "NotFound",
					Description: "given binding_id was not found for this instance_id",
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
		}
	}

	binding, exists := deployment.GetBinding(bindingID)
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given binding_id was not found for this instance_id",
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
		// to do: create response from operation fields ;)
	}

	//DIFFERENT APPROACH ???!
	//USING INSTANCEOPERATIONPOLLRESPONSE INSTEAD OF BINDINGOPERATIONPOLLRESPONSE
	//-> FIELDS FOR INSTANCE_USABLE AND UPDATE_REPEATABLE CAN SIMPLY BE OMITTED
	var responseDescription *string
	if bindingService.settings.BindingSettings.ReturnDescriptionLastOperation {
		description := "Default description"
		responseDescription = &description
	}
	pollResponse := model.InstanceOperationPollResponse{
		State:       *operation.State(),
		Description: responseDescription,
	}
	statusCode := 200 //ok
	if operation.InstanceUsable() != nil && !*operation.InstanceUsable() && operation.SupposedToFail() {
		statusCode = 410 //gone
	}
	return statusCode, &pollResponse, nil
}

func (bindingService *BindingService) Unbind(deleteRequest *model.DeleteRequest, instanceID *string, bindingID *string,
	serviceID *string, planID *string) (int, *string, *model.ServiceBrokerError) {
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(deleteRequest.Parameters)
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 410, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	binding, exists := deployment.GetBinding(bindingID)
	if !exists {
		return 410, nil, &model.ServiceBrokerError{
			Error:       "Gone",
			Description: "service binding does not exist",
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

	//CRITICAL STUFF REMOVED HERE????!!!
	//do stuff...
	//remove from binding map
	var operationID *string
	if !*requestSettings.FailAtOperation { //not fail = success lol
		deployment.RemoveBinding(bindingID)
		delete(*bindingService.bindingInstances, *bindingID)
		operationID = binding.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, &bindingService.lastOperationOfDeletedBinding, bindingID)
	} else {
		operationID = binding.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, nil, nil)
	}

	//(*bindingService.bindingInstances)[*bindingID] = nil

	//add operation to graveyard - done by adding parameter to DoOperation
	/*async := false
	if requestSettings.AsyncEndpoint != nil {
		async = *requestSettings.AsyncEndpoint
	}

	*/
	//shouldFail := false
	//bindingService.lastOperationOfDeletedBinding[*bindingID] =

	var response string
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync {
			response = *operationID
			//this is still ok, the response in the binding is only used when not async created or when fetched
		}
		return 202, &response, nil
	}

	/*
		changes to poll operation:
		when a service/binding is deleted, it is removed from the map
		therefore it is useless to check the state of an async deletion (as soon as it is deleted, it is not found in the map)
		solution: graveyard map[string]operation:
		key = serviceID/bindingID, value = LAST operation
		saves the LAST operation which is caused by the deletion
		poll operation. binding not found in map of bindings? look in graveyard:
		operation.async -> return 410 gone operation
		!operation.async -> return 200 ok operation (with state = failed. this represents the creation/rotation)
	*/
	response = "{}"
	return 200, &response, nil
}

func (bindingService *BindingService) CurrentBindings() *map[string]*model.ServiceBinding {
	return bindingService.bindingInstances
}
