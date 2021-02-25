package model

import (
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"log"
	"strconv"
	"time"
)

type ServiceDeployment struct {
	serviceID     string
	planID        string
	instanceID    string
	parameters    *interface{}
	dashboardURL  *string
	metadata      *ServiceInstanceMetadata
	lastOperation *Operation
	operations    map[string]*Operation
	/*
		Alternative solution ???
		requestIDToOperation map[string]map[string]*Operation?? Seems too sketchy/too much???
		originIDToOperation instead/too???
	*/
	requestIDToOperation map[string]*Operation
	nextOperationNumber  int

	//use
	//this to indicate if still usable
	state string
	//and
	//this to indicate if update running
	updatingOperations map[string]bool //if an operation is updating the instance, its name is in this string, otherwise nil
	/*
		if fetching instance and updatingOperation != nil: get the state of the Operation (by using the name from updatingOperation)
		set updatingOperation to nil and continue (return true - instance fetchable), if Operation is finished. otherwise
		leave updatingOperation as is and return (return false - instance not fetchable?)
	*/

	async                    bool
	secondsToFinishOperation int
	organizationID           *string
	spaceID                  *string
	doOperationChan          chan int
	deploymentUsable         bool
	//lastOperation
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

func (serviceDeployment *ServiceDeployment) PlanID() string {
	return serviceDeployment.planID
}

func (serviceDeployment *ServiceDeployment) ServiceID() string {
	return serviceDeployment.serviceID
}

func (serviceDeployment *ServiceDeployment) State() string {
	return serviceDeployment.state
}

func (serviceDeployment *ServiceDeployment) Metadata() *ServiceInstanceMetadata {
	return serviceDeployment.metadata
}

func (serviceDeployment *ServiceDeployment) DashboardURL() *string {
	return serviceDeployment.dashboardURL
}

func (serviceDeployment *ServiceDeployment) LastOperationID() *string {
	return serviceDeployment.lastOperation.Name()
}

/*var nextOperationNumber chan int

func init() {
	nextOperationNumber = make(chan int, 1)
	log.Println("before sending to channel")
	nextOperationNumber <- 0
	log.Println("after sending to channel")
}*/

func (serviceDeployment *ServiceDeployment) Parameters() *interface{} {
	if serviceDeployment.parameters == nil {
		return nil
	}
	return serviceDeployment.parameters
}
func NewServiceDeployment(instanceID string, provisionRequest *ProvideServiceInstanceRequest, settings *Settings) (*ServiceDeployment, *string) {
	serviceDeployment := ServiceDeployment{
		serviceID:      provisionRequest.ServiceID,
		planID:         provisionRequest.PlanID,
		instanceID:     instanceID,
		parameters:     provisionRequest.Parameters,
		organizationID: &provisionRequest.OrganizationGUID,
		spaceID:        &provisionRequest.SpaceGUID,
		operations:     make(map[string]*Operation),
		//async:                    async,
		//secondsToFinishOperation: settings.ProvisionSettings.SecondsToFinish,
		//lastOperation: "task_0",
		nextOperationNumber: 0,
		updatingOperations:  make(map[string]bool),
		doOperationChan:     make(chan int, 1),
		deploymentUsable:    true,
	}
	var requestSettings *RequestSettings
	requestSettings, _ = GetRequestSettings(provisionRequest.Parameters)

	/*var opNumber int
	opNumber = <-nextOperationNumber
	serviceDeployment.lastOperation = "task_" + strconv.Itoa(opNumber)
	nextOperationNumber <- opNumber + 1

	*/
	//CHECK IF OPERATION SHOULD BE ALSO DONE WITH SYNC!!! PROBABLY?!
	/*if !(!settings.ProvisionSettings.Async && !settings.ProvisionSettings.Operation) {
		serviceDeployment.doOperation(settings.ProvisionSettings.SecondsToFinish)
	}

	*/
	if settings.ProvisionSettings.DashboardURL {
		serviceDeployment.buildDashboardURL()
	}
	if settings.ProvisionSettings.Metadata {
		serviceDeployment.metadata = &ServiceInstanceMetadata{
			Labels: map[string]string{
				"labelKey": "labelValue",
			},
			Attributes: map[string]string{
				"attributesKey": "attributesValue",
			},
		}
	}

	/*if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint { //settings.ProvisionSettings.Async {
		//in progress here or in deploy()? if it's here then the service will safely have a state when returned
		//operation := NewOperation()
		//serviceDeployment.state = "in progress"
		//go serviceDeployment.deploy(settings.ProvisionSettings.SecondsToFinish)
		//provisioning "never" fails / is not specified in the specs???
		serviceDeployment.doOperation(*requestSettings.SecondsToComplete, false, nil)
	} else {
		//provisioning "never" fails / is not specified in the specs?
		serviceDeployment.doOperation(*requestSettings.SecondsToComplete, false, nil)
		//hier einfach sleepen bevor returned wird??????
	}

	*/
	shouldFail := false
	operationID := serviceDeployment.doOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, &shouldFail, nil, nil)
	log.Println("here comes the deployment")
	log.Println(serviceDeployment)
	log.Println(*serviceDeployment.dashboardURL)
	//ATTENTION?!
	//WRITE HERE IN CHANNEL FROM WHICH WILL BE CONSUMED WHEN DELETING???
	//ATTENTION?!
	//DOES DELETING NEED TO BLOCKED? PROBABLY YES, BECAUSE INSTANCE ID IS KNOWN BY PLATFORM (PLATFORM PROVIDES ID)
	return &serviceDeployment, operationID
}

//check return types
/*
func (serviceDeployment *ServiceDeployment) GetStateOfOperation() string {

}

*/

func (serviceDeployment *ServiceDeployment) Update(updateServiceInstanceRequest *UpdateServiceInstanceRequest,
	settings *Settings) (*string, *ServiceBrokerError) { //}, requestSettings *RequestSettings)  {

	//check even before if instance usable?!
	//not here? the spec is saying that fetching while updating is forbidden, not updating while updating (which will be a different issue lol)???
	/*
		if serviceDeployment.updatingOperation != nil && serviceDeployment.operations[*serviceDeployment.updatingOperation] != nil {
			if state := serviceDeployment.operations[*serviceDeployment.updatingOperation].State(); *state != PROGRESSING {

			}
		}

	*/

	//could also be passed instead
	requestSettings, _ := GetRequestSettings(updateServiceInstanceRequest.Parameters)
	//make use of context or ignore????
	//change ONLY parameters and planid???
	if !*requestSettings.FailAtOperation {
		if updateServiceInstanceRequest.PlanId != nil {
			fmt.Println("plan id prior to update: " + serviceDeployment.planID)
			serviceDeployment.planID = *updateServiceInstanceRequest.PlanId
			fmt.Println("plan id after update: " + serviceDeployment.planID)
		}
		if updateServiceInstanceRequest.Parameters != nil {
			log.Println("trying to change parameters, old :")
			log.Println(serviceDeployment.parameters)
			//oldParam :=
			serviceDeployment.parameters = updateServiceInstanceRequest.Parameters
			log.Println("new: ")
			log.Println(serviceDeployment.parameters)
		}
	}

	operationID := serviceDeployment.doOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete,
		requestSettings.FailAtOperation, requestSettings.UpdateRepeatableAfterFail,
		requestSettings.InstanceUsableAfterFail)
	//setting updatingOperaton field to indicate ongoing update
	//serviceDeployment.updatingOperations[*operationID] = true
	//serviceDeployment.updatingOperation = operationID
	return operationID, nil
}

/*func (serviceDeployment *ServiceDeployment) deploy(seconds int) {

		77milliSecondsToFinishOperation warten, dann state auf succeeded

	time.Sleep(time.duration(seconds) * time.Second)
	serviceDeployment.state = "succeeded"
}*/

func (serviceDeployment *ServiceDeployment) UpdatesRunning() bool {
	//will entry be removed if value is set to false? how fast will this use up memory if not???
	for operationName, running := range serviceDeployment.updatingOperations {
		if running {
			//state := serviceDeployment.operations[operationName].State()
			if *serviceDeployment.operations[operationName].State() == PROGRESSING {
				return true
			}
			serviceDeployment.updatingOperations[operationName] = false

		}
	}
	return false
}

func (serviceDeployment *ServiceDeployment) doOperation(async bool, duration int, shouldFail *bool, updateRepeatable *bool, deploymentUsable *bool) *string {
	serviceDeployment.doOperationChan <- 1
	operationID := "task_" + strconv.Itoa(serviceDeployment.nextOperationNumber)
	serviceDeployment.updatingOperations[operationID] = true
	operation := NewOperation(operationID, float64(duration), *shouldFail, updateRepeatable, deploymentUsable)
	serviceDeployment.lastOperation = operation
	serviceDeployment.operations[operationID] = operation
	fmt.Printf("nextoperationnumber before increment %v\n", serviceDeployment.nextOperationNumber)
	serviceDeployment.nextOperationNumber++ // = serviceDeployment.nextOperationNumber + 1
	fmt.Printf("nextoperationnumber after increment %v\n", serviceDeployment.nextOperationNumber)
	fmt.Println("operations:")
	fmt.Println(serviceDeployment.operations)
	<-serviceDeployment.doOperationChan
	//WAIT HERE IF !ASYNC???!
	if !async {
		//CHECK=!
		time.Sleep(time.Duration(duration) * time.Second)
	}
	if deploymentUsable != nil && *shouldFail && !*deploymentUsable {
		serviceDeployment.deploymentUsable = *deploymentUsable
	}
	//async check necessary??? checking now somewhere else if async (and therefore if operationID should be in response)
	return &operationID
	//fmt.Println(serviceDeployment.operations)
	/*operation := "task_" + strconv.Itoa(serviceDeployment.nextOperationNumber)
	serviceDeployment.lastOperation = &operation
	serviceDeployment.nextOperationNumber++*/
}

func (serviceDeployment *ServiceDeployment) buildDashboardURL() {
	url := "http://" + generator.RandomString(4) + ".com/" + generator.RandomString(4)
	serviceDeployment.dashboardURL = &url
}
