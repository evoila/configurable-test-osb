package model

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"log"
	"strconv"
	"time"
)

type ServiceDeployment struct {
	serviceID                string
	planID                   string
	instanceID               string
	parameters               interface{}
	dashboardURL             *string
	metadata                 *ServiceInstanceMetadata
	operation                *string
	nextOperationNumber      int
	state                    string
	async                    bool
	secondsToFinishOperation int
	//operation
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

func (serviceDeployment *ServiceDeployment) Operation() *string {
	return serviceDeployment.operation
}

/*var nextOperationNumber chan int

func init() {
	nextOperationNumber = make(chan int, 1)
	log.Println("before sending to channel")
	nextOperationNumber <- 0
	log.Println("after sending to channel")
}*/

func (serviceDeployment *ServiceDeployment) Parameters() *interface{} {
	return &serviceDeployment.parameters
}

func NewServiceDeployment(instanceID string, provisionRequest *ProvideServiceInstanceRequest, settings *Settings) *ServiceDeployment {
	serviceDeployment := ServiceDeployment{
		serviceID:  provisionRequest.ServiceID,
		planID:     provisionRequest.PlanID,
		instanceID: instanceID,
		parameters: provisionRequest.Parameters,
		//async:                    async,
		secondsToFinishOperation: settings.ProvisionSettings.SecondsToFinish,
		//operation: "task_0",
		nextOperationNumber: 0,
	}
	/*var opNumber int
	opNumber = <-nextOperationNumber
	serviceDeployment.operation = "task_" + strconv.Itoa(opNumber)
	nextOperationNumber <- opNumber + 1

	*/
	if !(!settings.ProvisionSettings.Async && !settings.ProvisionSettings.Operation) {
		serviceDeployment.doOperation()
	}
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
	if settings.ProvisionSettings.Async {
		//in progress here or in deploy()? if it's here then the service will safely have a state when returned
		serviceDeployment.state = "in progress"
		go serviceDeployment.deploy()
	} else {
		//state notwending, wenn !async? mal versuchen bei !async auf state zu verzichten
		serviceDeployment.state = "succeeded"
	}
	log.Println("here comes the deployment")
	log.Println(serviceDeployment)
	log.Println(*serviceDeployment.dashboardURL)
	return &serviceDeployment
}

func (serviceDeployment *ServiceDeployment) deploy() {
	/*
		milliSecondsToFinishOperation warten, dann state auf succeeded
	*/
	time.Sleep(time.Duration(serviceDeployment.secondsToFinishOperation) * time.Second)
	serviceDeployment.state = "succeeded"
}

func (serviceDeployment *ServiceDeployment) doOperation() {
	operation := "task_" + strconv.Itoa(serviceDeployment.nextOperationNumber)
	serviceDeployment.operation = &operation
	serviceDeployment.nextOperationNumber++
}

func (serviceDeployment *ServiceDeployment) buildDashboardURL() {
	url := "http://" + generator.RandomString(4) + ".com/" + generator.RandomString(4)
	serviceDeployment.dashboardURL = &url
}
