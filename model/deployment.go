package model

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"log"
	"strconv"
	"time"
)

type ServiceDeployment struct {
	serviceID            *string
	planID               *string
	instanceID           string
	parameters           *interface{}
	dashboardURL         *string
	metadata             *ServiceInstanceMetadata
	bindings             map[string]*ServiceBinding
	deletedBindings      []string
	lastOperation        *Operation
	operations           map[string]*Operation
	requestIDToOperation map[string]*Operation
	nextOperationNumber  int
	//indicates if update(s) running
	updatingOperations       map[string]bool
	async                    bool
	secondsToFinishOperation int
	organizationID           *string
	spaceID                  *string
	doOperationChan          chan int
	deploymentUsable         bool
	fetchResponse            *FetchingServiceInstanceResponse
	settings                 *Settings
	catalog                  *Catalog
}

func (serviceDeployment *ServiceDeployment) BindingDeleted(bindingToFind *string) bool {
	for _, bindingID := range serviceDeployment.deletedBindings {
		if bindingID == *bindingToFind {
			return true
		}
	}
	return false
}

func (serviceDeployment *ServiceDeployment) FetchResponse() *FetchingServiceInstanceResponse {
	return serviceDeployment.fetchResponse
}

func (serviceDeployment *ServiceDeployment) DeploymentUsable() bool {
	return serviceDeployment.deploymentUsable
}

func (serviceDeployment *ServiceDeployment) GetLastOperation() *Operation {
	return serviceDeployment.lastOperation
}

func (serviceDeployment *ServiceDeployment) GetOperationByName(operationName string) *Operation {
	return serviceDeployment.operations[operationName]
}

func (serviceDeployment *ServiceDeployment) SpaceID() *string {
	return serviceDeployment.spaceID
}

func (serviceDeployment *ServiceDeployment) OrganizationID() *string {
	return serviceDeployment.organizationID
}

func (serviceDeployment *ServiceDeployment) PlanID() *string {
	return serviceDeployment.planID
}

func (serviceDeployment *ServiceDeployment) ServiceID() *string {
	return serviceDeployment.serviceID
}

/*func (serviceDeployment *ServiceDeployment) State() string {
	return serviceDeployment.state
}

*/

func (serviceDeployment *ServiceDeployment) Metadata() *ServiceInstanceMetadata {
	return serviceDeployment.metadata
}

func (serviceDeployment *ServiceDeployment) DashboardURL() *string {
	return serviceDeployment.dashboardURL
}

func (serviceDeployment *ServiceDeployment) LastOperationID() *string {
	return serviceDeployment.lastOperation.Name()
}

func (serviceDeployment *ServiceDeployment) Parameters() *interface{} {
	if serviceDeployment.parameters == nil {
		return nil
	}
	return serviceDeployment.parameters
}
func NewServiceDeployment(instanceID string, provisionRequest *ProvideServiceInstanceRequest, settings *Settings,
	catalog *Catalog) (*ServiceDeployment, *string) {
	serviceDeployment := ServiceDeployment{
		serviceID:           &provisionRequest.ServiceID,
		planID:              &provisionRequest.PlanID,
		instanceID:          instanceID,
		parameters:          provisionRequest.Parameters,
		organizationID:      &provisionRequest.OrganizationGUID,
		spaceID:             &provisionRequest.SpaceGUID,
		bindings:            make(map[string]*ServiceBinding),
		operations:          make(map[string]*Operation),
		nextOperationNumber: 0,
		updatingOperations:  make(map[string]bool),
		doOperationChan:     make(chan int, 1),
		deploymentUsable:    true,
		//fetchResponse:       &FetchingServiceInstanceResponse{},
		settings:        settings,
		catalog:         catalog,
		deletedBindings: make([]string, 0),
	}
	var requestSettings *RequestSettings
	requestSettings, _ = GetRequestSettings(provisionRequest.Parameters)
	if settings.ProvisionSettings.CreateDashboardURL {
		serviceDeployment.buildDashboardURL()
	}
	if settings.ProvisionSettings.CreateMetadata {
		serviceDeployment.metadata = &ServiceInstanceMetadata{
			Labels: map[string]string{
				"labelKey": "labelValue",
			},
			Attributes: map[string]string{
				"attributesKey": "attributesValue",
			},
		}
	}
	offering, _ := catalog.GetServiceOfferingById(provisionRequest.ServiceID)
	if *offering.InstancesRetrievable {
		serviceDeployment.setResponse()
	}
	operationID := serviceDeployment.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, nil, nil, nil, nil, true)
	return &serviceDeployment, operationID
}

//right now, updating while an update is running is allowed
func (serviceDeployment *ServiceDeployment) Update(updateServiceInstanceRequest *UpdateServiceInstanceRequest) (*string, *ServiceBrokerError) {
	requestSettings, _ := GetRequestSettings(updateServiceInstanceRequest.Parameters)
	//change ONLY parameters and planid???
	if !*requestSettings.FailAtOperation {
		if updateServiceInstanceRequest.PlanId != nil {
			serviceDeployment.planID = updateServiceInstanceRequest.PlanId
		}
		if updateServiceInstanceRequest.Parameters != nil {
			serviceDeployment.parameters = updateServiceInstanceRequest.Parameters
		}
	}
	operationID := serviceDeployment.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, requestSettings.FailAtOperation, requestSettings.UpdateRepeatableAfterFail, requestSettings.InstanceUsableAfterFail, nil, nil, true)
	return operationID, nil
}

func (serviceDeployment *ServiceDeployment) Blocked() bool {
	/*
		entry will now be removed if state != progressing
		-> true false check not necessary? only look, if entry in slice (instead of a map)???
		OR: change updatingoperations to map[string]*Operation???!!! this sounds good
	*/
	for operationName, running := range serviceDeployment.updatingOperations {
		if running {
			if serviceDeployment.operations[operationName] != nil && *serviceDeployment.operations[operationName].State() == PROGRESSING {
				return true
			}
			delete(serviceDeployment.updatingOperations, operationName)
			serviceDeployment.updatingOperations[operationName] = false

		}
	}
	return false
}

func (serviceDeployment *ServiceDeployment) DoOperation(async bool, duration int, shouldFail *bool,
	updateRepeatable *bool, deploymentUsable *bool, lastOperationOfDeletedInstance *map[string]*Operation, id *string, blocked bool) *string {
	serviceDeployment.doOperationChan <- 1
	operationID := "task_" + strconv.Itoa(serviceDeployment.nextOperationNumber)
	if blocked {
		serviceDeployment.updatingOperations[operationID] = true
	}
	operation := NewOperation(operationID, float64(duration), *shouldFail, updateRepeatable, deploymentUsable, async)
	if lastOperationOfDeletedInstance != nil && id != nil {
		(*lastOperationOfDeletedInstance)[*id] = operation
	}
	serviceDeployment.lastOperation = operation
	serviceDeployment.operations[operationID] = operation
	serviceDeployment.nextOperationNumber++
	<-serviceDeployment.doOperationChan
	if !async {
		time.Sleep(time.Duration(duration) * time.Second)
	}
	if deploymentUsable != nil && *shouldFail && !*deploymentUsable {
		serviceDeployment.deploymentUsable = *deploymentUsable
	}
	return &operationID
}

func (serviceDeployment *ServiceDeployment) AddBinding(serviceBinding *ServiceBinding) {
	if serviceBinding != nil {
		serviceDeployment.bindings[*serviceBinding.bindingID] = serviceBinding
	} else {
		log.Println("Nil pointer passed when adding binding, this should not have happened.")
	}
}

func (serviceDeployment *ServiceDeployment) GetBinding(bindingID *string) (*ServiceBinding, bool) {
	serviceBinding, exists := (serviceDeployment.bindings)[*bindingID]
	return serviceBinding, exists
}

func (serviceDeployment *ServiceDeployment) RemoveBinding(bindingID *string) {
	binding := serviceDeployment.bindings[*bindingID]
	serviceDeployment.deletedBindings = append(serviceDeployment.deletedBindings, *binding.bindingID)
	delete(serviceDeployment.bindings, *bindingID)
}

func (serviceDeployment *ServiceDeployment) AmountOfBindings() int {
	return len(serviceDeployment.bindings)
}

func (serviceDeployment *ServiceDeployment) buildDashboardURL() {
	url := "http://" + generator.RandomString(4) + ".com/" + generator.RandomString(4)
	serviceDeployment.dashboardURL = &url
}

func (serviceDeployment *ServiceDeployment) setResponse() {
	serviceDeployment.fetchResponse = &FetchingServiceInstanceResponse{}
	if serviceDeployment.settings.FetchServiceInstanceSettings.ReturnServiceID {
		serviceDeployment.fetchResponse.ServiceId = serviceDeployment.serviceID
	}
	if serviceDeployment.settings.FetchServiceInstanceSettings.ReturnPlanID {
		serviceDeployment.fetchResponse.PlanId = serviceDeployment.planID
	}
	if serviceDeployment.settings.FetchServiceInstanceSettings.ReturnDashboardURL {
		serviceDeployment.fetchResponse.DashboardUrl = serviceDeployment.dashboardURL
	}
	if serviceDeployment.settings.FetchServiceInstanceSettings.ReturnParameters {
		serviceDeployment.fetchResponse.Parameters = serviceDeployment.parameters
	}
	if serviceDeployment.settings.FetchServiceInstanceSettings.ReturnMaintenanceInfo {
		serviceOffering, _ := serviceDeployment.catalog.GetServiceOfferingById(*serviceDeployment.serviceID)
		servicePlan, _ := serviceOffering.GetPlanByID(*serviceDeployment.planID)
		serviceDeployment.fetchResponse.MaintenanceInfo = servicePlan.MaintenanceInfo
	}
	if serviceDeployment.settings.FetchServiceInstanceSettings.ReturnMetadata {
		serviceDeployment.fetchResponse.Metadata = serviceDeployment.metadata
	}
}
