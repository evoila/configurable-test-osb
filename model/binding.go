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
	lastOperation       *Operation
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

//serviceBinding.SetInformationReturned marks that the binding information has been returned in case the information
//should only be returned once
func (serviceBinding *ServiceBinding) SetInformationReturned(informationReturned bool) {
	serviceBinding.informationReturned = informationReturned
}

func (serviceBinding *ServiceBinding) Response() *CreateRotateFetchBindingResponse {
	return serviceBinding.response
}

func (serviceBinding *ServiceBinding) GetOperationByName(operationName string) *Operation {
	return serviceBinding.operations[operationName]
}

func (serviceBinding *ServiceBinding) GetLastOperation() *Operation {
	return serviceBinding.lastOperation
}

//NewServiceBinding creates a new Service Binding and adds the binding to the map of existing Bindings and to the
//corresponding deployment
//Returns a pointer to the new binding and the operationID in case the endpoint is called async
func NewServiceBinding(bindingID *string, bindingRequest *CreateBindingRequest, settings *Settings,
	catalog *Catalog, bindingInstances *map[string]*ServiceBinding, deployment *ServiceDeployment) (*ServiceBinding, *string) {
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
		response:            &CreateRotateFetchBindingResponse{},
	}
	deployment.AddBinding(&serviceBinding)
	serviceBinding.serviceOffering, _ = catalog.GetServiceOfferingById(*bindingRequest.ServiceID)
	serviceBinding.setResponse()
	(*bindingInstances)[*bindingID] = &serviceBinding
	var requestSettings *RequestSettings
	requestSettings, _ = GetRequestSettings(bindingRequest.Parameters)
	shouldFail := false
	operationID := serviceBinding.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, &shouldFail, nil, nil)
	return &serviceBinding, operationID
}

//is this ok??? what happens, if the original service binding is deleted??? is the garbage collector "smart" enough to
//store the field of the new serviceBinding separately even though it is created from the old one?
func (serviceBinding *ServiceBinding) RotateBinding(rotateBindingRequest *RotateBindingRequest, deployment *ServiceDeployment) (*ServiceBinding, *string) {
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
	deployment.AddBinding(&newServiceBinding)
	var requestSettings *RequestSettings
	requestSettings, _ = GetRequestSettings(rotateBindingRequest.Parameters)
	shouldFail := false
	operationID := newServiceBinding.DoOperation(*requestSettings.AsyncEndpoint, *requestSettings.SecondsToComplete, &shouldFail, nil, nil)
	return &newServiceBinding, operationID
}

func (serviceBinding *ServiceBinding) DoOperation(async bool, duration int, shouldFail *bool, lastOperationOfDeletedInstance *map[string]*Operation, id *string) *string {
	serviceBinding.doOperationChan <- 1
	operationID := "task_" + strconv.Itoa(serviceBinding.nextOperationNumber)
	operation := NewOperation(operationID, float64(duration), *shouldFail, nil, nil, async)
	if lastOperationOfDeletedInstance != nil && id != nil {
		(*lastOperationOfDeletedInstance)[*id] = operation
	}
	serviceBinding.lastOperation = operation
	serviceBinding.operations[operationID] = operation
	serviceBinding.nextOperationNumber++ // = serviceBinding.nextOperationNumber + 1
	<-serviceBinding.doOperationChan
	//WAIT HERE IF !ASYNC???!
	if !async {
		time.Sleep(time.Duration(duration) * time.Second)
	}
	return &operationID
}

func (serviceBinding *ServiceBinding) setResponse() {

	if serviceBinding.settings.BindingSettings.BindingMetadataSettings.ReturnMetadata {
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
		var credentials interface{}

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
		endpoint := Endpoint{
			Host:  "myHost",
			Ports: []string{"1234", "5678"},
		}
		if serviceBinding.settings.BindingSettings.BindingEndpointSettings.ReturnProtocol {
			endpoint.Protocol = &serviceBinding.settings.BindingSettings.BindingEndpointSettings.ProtocolValue
		}
		endpoints := []Endpoint{endpoint}
		serviceBinding.response.Endpoints = &endpoints
	}
	if serviceBinding.settings.BindingSettings.ReturnParameters {
		serviceBinding.response.Parameters = serviceBinding.parameters

	}
}
