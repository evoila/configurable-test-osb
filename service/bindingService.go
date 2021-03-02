package service

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/google/go-cmp/cmp"
	"log"
)

type BindingService struct {
	//POINTER TO SERVICE INSTANCES/DEPLOYMENTS???? PRETTY SURE YES
	serviceInstances *map[string]*model.ServiceDeployment
	//key = binding_id
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

func (bindingService *BindingService) CreateBinding(bindingRequest *model.CreateBindingRequest, instanceID *string,
	bindingID *string) (int, *model.CreateRotateFetchBindingResponse, *model.ServiceBrokerError) {
	deployment, exists := (*bindingService.serviceInstances)[*instanceID]
	if !exists {
		return 404, nil, &model.ServiceBrokerError{
			Error:       "NotFound",
			Description: "given instance_id was not found",
		}
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(bindingRequest.Parameters)
	if binding, exists := (*bindingService.bindingInstances)[*bindingID]; exists == true {
		if bindingService.settings.BindingSettings.StatusCodeOK && !*requestSettings.AsyncEndpoint {
			log.Println("binding with given id exists. now comparing equality of requested binding and existing binding")
			if cmp.Equal(bindingRequest.Parameters, binding.Parameters()) &&
				cmp.Equal(bindingRequest.Context, binding.Context()) &&
				*bindingRequest.ServiceID == *binding.ServiceID() &&
				*bindingRequest.PlanID == *binding.PlanID() &&
				cmp.Equal(bindingRequest.AppGUID, binding.AppGuid()) &&
				cmp.Equal(bindingRequest.BindResource, binding.BindResource()) {
				log.Println("requested binding and existing binding are identical")
				if bindingService.settings.BindingSettings.ReturnBindingInformationOnce && binding.InformationReturned() {
					return 200, &model.CreateRotateFetchBindingResponse{}, nil
				}
				response := binding.Response()
				response.Parameters = nil
				return 200, response, nil
			}
			log.Println("requested binding and existing binding are different")
			if bindingRequest.BindResource != nil && binding.BindResource() != nil {
				log.Println(*bindingRequest.BindResource, *binding.BindResource())
			} else {
				log.Println("bindResource in request or in existing binding is nil")
			}

		}

		return 409, nil, &model.ServiceBrokerError{
			Error:       "InstanceIDConflict",
			Description: "The given binding_id is already in use",
		}
	}
	if *bindingRequest.ServiceID != deployment.ServiceID() {
		//log.Println("this should not be here...")
		return 400, nil, &model.ServiceBrokerError{
			Error:       "ServiceIDMatch",
			Description: "The given service_id does not match the service_id of the instance",
		}
	}
	if *bindingRequest.PlanID != deployment.PlanID() {
		return 400, nil, &model.ServiceBrokerError{
			Error:       "PlanIDMatch",
			Description: "The given plan_id does not match the plan_id of the instance",
		}
	}
	/*shouldFail := false
	operationID := newServiceBinding.doOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, &shouldFail, nil, nil)*/
	//var binding *model.ServiceBinding
	/*binding := &model.ServiceBinding{bindingID: bindingID}

	 */
	//deployment.AddBinding(binding)
	binding, operationID := model.NewServiceBinding(bindingID, bindingRequest, bindingService.settings, bindingService.catalog, bindingService.bindingInstances, deployment)
	//this is now done in model.NewServiceBinding)
	//(*bindingService.bindingInstances)[*bindingID] = binding

	//is just using var ok here or do i actually need to create CrateRotateBindingResponse{} ???
	//response.Operation = "hello"

	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		var response model.CreateRotateFetchBindingResponse
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync {
			response.Operation = operationID
			//this is still ok, the response in the binding is only used when not async created or when fetched
		}
		return 202, &response, nil
	}
	//CREATE OWN FUNCTION???! THIS RESPONSE IS ALSO CREATED WHEN THE BINDING IS FETCHED
	//ONLY DIFFERENCE IS, THAT FETCHING CAN ALSO RETURN PARAMETERS (which could be changed to nil here afterwards)
	/*if bindingService.settings.BindingSettings.BindingMetadataSettings.ReturnMetadata{
		//make metadata
	}
	if bindingService.settings.BindingSettings.ReturnCredentials {
		//make credentials
	}
	if bindingService.settings.BindingSettings.ReturnSyslogDrainURL{

	}

	*/
	binding.SetInformationReturned(true)
	//should not be necessary??? when parameters shall be returned by fetch, they are set to nil afterwards
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
	//response.Operation = "hello"
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
	log.Println("binding \"exists\":")
	log.Println(binding)
	log.Println(*binding)
	if serviceID != nil && *serviceID != "" {
		if deployment.ServiceID() != *serviceID {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "ServiceIDMatch",
				Description: "The given service_id does not match the service_id of the instance",
			}
		}
	}
	if planID != nil && *planID != "" {
		if deployment.PlanID() != *planID {
			return 400, nil, &model.ServiceBrokerError{
				Error:       "ServiceIDMatch",
				Description: "The given plan_id does not match the plan_id of the instance",
			}
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

func (bindingService *BindingService) PollOperationState(instanceID *string, bindingID *string, serviceID *string, planID *string, operationName *string) (int, *model.InstanceOperationPollResponse, *model.ServiceBrokerError) {
	var deployment *model.ServiceDeployment
	var exists bool
	deployment, exists = (*bindingService.serviceInstances)[*instanceID]
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
			Description: "given binding_id was not found for this instance_id",
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
		operation = binding.GetOperationByName(*operationName)
		if operation == nil {
			return 404, nil, &model.ServiceBrokerError{
				Error:       "OperationID",
				Description: "The given operation does not exist for the service instance",
			}
		}
	} else {
		operation = binding.GetLastOperation()
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
		return 404, nil, &model.ServiceBrokerError{
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

	//CRITICAL STUFF REMOVED HERE????!!!
	//do stuff...
	//remove from binding map
	deployment.RemoveBinding(bindingID)
	delete(*bindingService.bindingInstances, *bindingID)
	//(*bindingService.bindingInstances)[*bindingID] = nil

	//add operation to graveyard - done by adding parameter to DoOperation
	async := false
	if requestSettings.AsyncEndpoint != nil {
		async = *requestSettings.AsyncEndpoint
	}
	shouldFail := false
	//bindingService.lastOperationOfDeletedBinding[*bindingID] =
	operationID := binding.DoOperation(async, *requestSettings.SecondsToComplete, &shouldFail, nil, nil, &bindingService.lastOperationOfDeletedBinding)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		var response string
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
		operation.async -> return 410 gone
		!operation.async -> return operation (with state = failed. this represents the creation/rotation)
	*/
	response := "{}"
	return 200, &response, nil
}

func (bindingService *BindingService) CurrentBindings() *map[string]*model.ServiceBinding {
	return bindingService.bindingInstances
}
