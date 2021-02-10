package model

import (
	"log"
	"strconv"
	"time"
)

type ServiceDeployment struct {
	instanceID               string
	parameters               interface{}
	operation                string
	state                    string
	async                    bool
	secondsToFinishOperation int

	//operation
}

var operationNumber chan int

func init() {
	operationNumber = make(chan int, 1)
	log.Println("before sending to channel")
	operationNumber <- 0
	log.Println("after sending to channel")
}

func (serviceDeployment *ServiceDeployment) Parameters() interface{} {
	return serviceDeployment.parameters
}

func NewServiceDeployment(instanceID string, parameters interface{}, async bool, timeToFinish int) *ServiceDeployment {
	serviceDeployment := ServiceDeployment{
		instanceID:               instanceID,
		parameters:               parameters,
		async:                    async,
		secondsToFinishOperation: timeToFinish,
	}
	var opNumber int
	opNumber = <-operationNumber
	serviceDeployment.operation = "task_" + strconv.Itoa(opNumber)
	operationNumber <- opNumber + 1
	if async {
		//in progress here or in deploy()? if it's here then the service will safely have a state when returned
		serviceDeployment.state = "in progress"
		go serviceDeployment.deploy()
	} else {
		//state notwending, wenn !async? mal versuchen bei !async auf state zu verzichten
		serviceDeployment.state = "succeeded"
	}
	log.Println("here comes the deployment")
	log.Println(serviceDeployment)
	return &serviceDeployment
}

func (serviceDeployment *ServiceDeployment) deploy() {
	/*
		milliSecondsToFinishOperation warten, dann state auf succeeded
	*/
	time.Sleep(time.Duration(serviceDeployment.secondsToFinishOperation) * time.Second)
	serviceDeployment.state = "succeeded"
}
