package service

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"log"
)

type BindingService struct {
	//POINTER TO SERVICE INSTANCES/DEPLOYMENTS???? PRETTY SURE YES
	serviceInstances *map[string]*model.ServiceDeployment
	//key = binding_id
	bindingInstances *map[string]*model.ServiceBinding
	settings         *model.Settings
	catalog          *model.Catalog
}

func NewBindingService(serviceInstances *map[string]*model.ServiceDeployment,
	bindingInstances *map[string]*model.ServiceBinding, settings *model.Settings, catalog *model.Catalog) *BindingService {
	return &BindingService{
		serviceInstances: serviceInstances,
		bindingInstances: bindingInstances,
		settings:         settings,
		catalog:          catalog,
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
	if _, exists := (*bindingService.bindingInstances)[*bindingID]; exists == true {
		return 409, nil, &model.ServiceBrokerError{
			Error:       "InstanceIDConflict",
			Description: "The given binding_id is already in use",
		}
	}
	if *bindingRequest.ServiceID != deployment.ServiceID() {
		log.Println("this should not be here...")
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

	binding, operationID := model.NewServiceBinding(*bindingID, bindingRequest, bindingService.settings, bindingService.catalog)
	(*bindingService.bindingInstances)[*bindingID] = binding

	//is just using var ok here or do i actually need to create CrateRotateBindingResponse{} ???
	var response model.CreateRotateFetchBindingResponse
	//response.Operation = "hello"
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(bindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
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
	return 201, binding.Response(), nil
}

func (bindingService *BindingService) RotateBinding(rotateBindingRequest *model.RotateBindingRequest, instanceID *string,
	bindingID *string) (int, *model.CreateRotateFetchBindingResponse, *model.ServiceBrokerError) {
	_, exists := (*bindingService.serviceInstances)[*instanceID]
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
	newBinding, operationID := oldBinding.RotateBinding(rotateBindingRequest)
	(*bindingService.bindingInstances)[*bindingID] = newBinding
	var response model.CreateRotateFetchBindingResponse
	//response.Operation = "hello"
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(rotateBindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint == true {
		if bindingService.settings.BindingSettings.ReturnOperationIfAsync {
			response.Operation = operationID
			//this is still ok, the response in the binding is only used when not async created or when fetched
		}
		return 202, &response, nil
	}
	//binding, operationID  := model.NewServiceBinding(*bindingID, rotateBindingRequest, bindingService.settings, bindingService.catalog)
	newBinding.SetInformationReturned(true)
	return 201, newBinding.Response(), nil
}

func (bindingService *BindingService) CurrentBindings() *map[string]*model.ServiceBinding {
	return bindingService.bindingInstances
}
