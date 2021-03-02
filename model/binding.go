package model

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"sort"
	"strconv"
	"time"
)

type ServiceBinding struct {
	bindingID           *string
	context             *interface{}
	serviceID           *string
	planID              *string
	appGuid             *string
	bindResource        *BindResource
	parameters          *interface{}
	operations          map[string]*Operation
	doOperationChan     chan int
	nextOperationNumber int

	lastOperation *Operation
	/*
		BindingMetadata *BindingMetadata
		Credentials *interface{}
		SyslogDrainUrl *string
		better idea:
	*/
	response            *CreateRotateFetchBindingResponse
	settings            *Settings
	catalog             *Catalog
	serviceOffering     *ServiceOffering
	informationReturned bool
}

func (serviceBinding *ServiceBinding) BindResource() *BindResource {
	return serviceBinding.bindResource
}

func (serviceBinding *ServiceBinding) AppGuid() *string {
	return serviceBinding.appGuid
}

func (serviceBinding *ServiceBinding) PlanID() *string {
	return serviceBinding.planID
}

func (serviceBinding *ServiceBinding) ServiceID() *string {
	return serviceBinding.serviceID
}

func (serviceBinding *ServiceBinding) Context() *interface{} {
	return serviceBinding.context
}

func (serviceBinding *ServiceBinding) Parameters() *interface{} {
	return serviceBinding.parameters
}

func (serviceBinding *ServiceBinding) BindingID() *string {
	return serviceBinding.bindingID
}

func (serviceBinding *ServiceBinding) InformationReturned() bool {
	return serviceBinding.informationReturned
}

func (serviceBinding *ServiceBinding) SetInformationReturned(informationReturned bool) {
	serviceBinding.informationReturned = informationReturned
}

func (serviceBinding *ServiceBinding) Response() *CreateRotateFetchBindingResponse {
	return serviceBinding.response
}

func NewServiceBinding(bindingID *string, bindingRequest *CreateBindingRequest, settings *Settings,
	catalog *Catalog) (*ServiceBinding, *string) {
	serviceBinding := ServiceBinding{
		bindingID:           bindingID,
		context:             bindingRequest.Context,
		serviceID:           bindingRequest.ServiceID,
		planID:              bindingRequest.PlanID,
		appGuid:             bindingRequest.AppGUID,
		bindResource:        bindingRequest.BindResource,
		parameters:          bindingRequest.Parameters,
		operations:          make(map[string]*Operation),
		doOperationChan:     make(chan int, 1),
		nextOperationNumber: 0,
		settings:            settings,
		catalog:             catalog,
		informationReturned: false,
		//serviceID: &bindingRequest.ServiceID,
		//planID: &bindingRequest.PlanID,
		response: &CreateRotateFetchBindingResponse{},
	}

	serviceBinding.serviceOffering, _ = catalog.GetServiceOfferingById(*bindingRequest.ServiceID)
	serviceBinding.setResponse()
	var requestSettings *RequestSettings
	requestSettings, _ = GetRequestSettings(bindingRequest.Parameters)
	shouldFail := false
	operationID := serviceBinding.doOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, &shouldFail, nil, nil)
	return &serviceBinding, operationID
}

//is this ok??? what happens, if the original service binding is deleted??? is the garbage collector "smart" enough to
//store the field of the new serviceBinding separately even though it is created from the old one?
func (serviceBinding *ServiceBinding) RotateBinding(rotateBindingRequest *RotateBindingRequest) (*ServiceBinding, *string) {
	newServiceBinding := ServiceBinding{
		context:             serviceBinding.context,
		appGuid:             serviceBinding.appGuid,
		bindResource:        serviceBinding.bindResource,
		parameters:          serviceBinding.parameters,
		operations:          make(map[string]*Operation),
		doOperationChan:     make(chan int, 1),
		nextOperationNumber: 0,
		//lastOperation:       nil,
		response:            serviceBinding.response,
		settings:            serviceBinding.settings,
		catalog:             serviceBinding.catalog,
		serviceOffering:     serviceBinding.serviceOffering,
		informationReturned: false,
	}
	var requestSettings *RequestSettings
	requestSettings, _ = GetRequestSettings(rotateBindingRequest.Parameters)
	shouldFail := false
	operationID := newServiceBinding.doOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, &shouldFail, nil, nil)
	return &newServiceBinding, operationID
}

func (serviceBinding *ServiceBinding) doOperation(async bool, duration int, shouldFail *bool, updateRepeatable *bool, deploymentUsable *bool) *string {
	serviceBinding.doOperationChan <- 1
	operationID := "task_" + strconv.Itoa(serviceBinding.nextOperationNumber)
	//serviceBinding.updatingOperations[operationID] = true
	//deploymentUsable not needed as the binding will always be "usable"???
	//updateRepeatable also not needed as bindings operations don't have a field update_repeatable
	operation := NewOperation(operationID, float64(duration), *shouldFail, nil, nil, async)
	serviceBinding.lastOperation = operation
	serviceBinding.operations[operationID] = operation
	//fmt.Printf("nextoperationnumber before increment %v\n", serviceBinding.nextOperationNumber)
	serviceBinding.nextOperationNumber++ // = serviceBinding.nextOperationNumber + 1
	//fmt.Printf("nextoperationnumber after increment %v\n", serviceBinding.nextOperationNumber)
	//fmt.Println("operations:")
	//fmt.Println(serviceBinding.operations)
	<-serviceBinding.doOperationChan
	//WAIT HERE IF !ASYNC???!
	if !async {
		//CHECK=!
		time.Sleep(time.Duration(duration) * time.Second)
	}
	//NOT NEEDED BECAUSE THE BINDING IS ALWAYS "USABLE"???
	/*if deploymentUsable != nil && *shouldFail && !*deploymentUsable {
		serviceBinding.deploymentUsable = *deploymentUsable
	}*/
	//async check necessary??? checking now somewhere else if async (and therefore if operationID should be in response)
	return &operationID
	//fmt.Println(serviceBinding.operations)
	/*operation := "task_" + strconv.Itoa(serviceBinding.nextOperationNumber)
	serviceBinding.lastOperation = &operation
	serviceBinding.nextOperationNumber++*/
}

func (serviceBinding *ServiceBinding) setResponse() {

	if serviceBinding.settings.BindingSettings.BindingMetadataSettings.ReturnMetadata {
		//make metadata
		metadata := BindingMetadata{}
		serviceBinding.response.Metadata = &metadata

		if serviceBinding.settings.BindingSettings.BindingMetadataSettings.ReturnExpiresAt {
			var expiresAt string
			expiresAt = time.Now().Add(240 * time.Hour).Format("2006-01-02T15:04:05-0700")
			serviceBinding.response.Metadata.ExpiresAt = &expiresAt
		}
		if serviceBinding.settings.BindingSettings.BindingMetadataSettings.ReturnRenewBefore {
			var renewBefore string
			renewBefore = time.Now().Add(240 * time.Hour).Format("2006-01-02T15:04:05-0700")
			serviceBinding.response.Metadata.RenewBefore = &renewBefore
		}
	}
	if serviceBinding.settings.BindingSettings.ReturnCredentials {
		//make credentials
		//credentials := interface
		var credentials interface{}

		/*content := "\"Uri\": \"mysql://mysqluser:pass@mysqlhost:3306/dbname\",\n    \"Username\": \"mysqluser\",\n    " +
			"\"Password\": \"pass\",\n    \"Host\": \"mysqlhost\",\n    \"Port\": 3306,\n    \"Database\": \"dbname\""
		if err := json.Unmarshal([]byte(content), &credentials); err != nil {
			log.Println("ERROR while creating credentials!")
			log.Println(err.Error())
		}

		*/
		credentials = struct {
			Uri      *string `json:"Uri"`
			Username string  `json:"Username"`
			Password string  `json:"Password"`
			Host     string  `json:"Host"`
			Port     int     `json:"Port"`
			Database string  `json:"Database"`
		}{
			Uri:      generator.RandomUriByFrequency("always", 6),
			Username: "user",
			Password: "userPassword",
			Host:     "Host",
			Port:     1234,
			Database: "myDatabase",
		}
		/*credentials = credStruct
		content, err := json.Marshal(credStruct)
		if err != nil {
			log.Println("ERROR while marshalling credStruct!")
			log.Println(err.Error())
		}
		if err = json.Unmarshal(content, &credentials); err != nil {
			log.Println("ERROR while creating credentials!")
			log.Println(err.Error())
		}

		*/
		serviceBinding.response.Credentials = &credentials
	}
	if serviceBinding.settings.BindingSettings.ReturnSyslogDrainURL && serviceBinding.serviceOffering.Requires != nil &&
		sort.SearchStrings(serviceBinding.serviceOffering.Requires, "syslog_drain") < len(serviceBinding.serviceOffering.Requires) {
		syslogDrainURL := generator.RandomUriByFrequency("always", 6)
		serviceBinding.response.SyslogDrainUrl = syslogDrainURL
	}
	if serviceBinding.settings.BindingSettings.ReturnRouteServiceURL && serviceBinding.serviceOffering.Requires != nil &&
		sort.SearchStrings(serviceBinding.serviceOffering.Requires, "route_forwarding") < len(serviceBinding.serviceOffering.Requires) {
		routeServiceURL := generator.RandomUriByFrequency("always", 6)
		serviceBinding.response.RouteServiceUrl = routeServiceURL
	}
	if serviceBinding.settings.BindingSettings.BindingVolumeMountSettings.ReturnVolumeMounts && serviceBinding.serviceOffering.Requires != nil &&
		sort.SearchStrings(serviceBinding.serviceOffering.Requires, "volume_mount") < len(serviceBinding.serviceOffering.Requires) {
		//make volume mounts
		volumeMount := VolumeMount{
			Driver:       "driverName",
			ContainerDir: "/hello/world",
			Mode:         "r",
			DeviceType:   "shared",
		}
		volumeMount.Device = &Device{
			VolumeId: generator.RandomString(8) + "-" + generator.RandomString(4) + "-" +
				generator.RandomString(4) + "-" + generator.RandomString(4) + "-" + generator.RandomString(12),
		}
		if serviceBinding.settings.BindingSettings.BindingVolumeMountSettings.ReturnMountConfig {
			var mountConfig interface{}
			mountConfig = struct {
				Key string `json:"key"`
			}{Key: "value"}
			volumeMount.Device.MountConfig = &mountConfig
		}
		volumeMounts := []VolumeMount{volumeMount}
		serviceBinding.response.VolumeMounts = &volumeMounts
	}
	if serviceBinding.settings.BindingSettings.BindingEndpointSettings.ReturnEndpoints {
		//make endpoints
		endpoint := Endpoint{
			Host:  "myHost",
			Ports: []string{"1234", "5678"},
		}
		if serviceBinding.settings.BindingSettings.BindingEndpointSettings.ReturnProtocol {
			//use protocol value in settings here
			endpoint.Protocol = &serviceBinding.settings.BindingSettings.BindingEndpointSettings.ProtocolValue
		}
		endpoints := []Endpoint{endpoint}
		serviceBinding.response.Endpoints = &endpoints
	}
	if serviceBinding.settings.BindingSettings.ReturnParameters {
		//set parameters pointer of response to parameter pointer of binding
		serviceBinding.response.Parameters = serviceBinding.parameters

	}
}
