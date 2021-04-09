package tests

import (
	"bytes"
	"encoding/json"
	"github.com/MaxFuhrich/configurable-test-osb/controller"
	"github.com/MaxFuhrich/configurable-test-osb/model"
	"github.com/MaxFuhrich/configurable-test-osb/server"
	"github.com/MaxFuhrich/configurable-test-osb/service"
	"github.com/gin-gonic/gin"
	"log"
	"sort"
	"testing"
)

func TestCreateBinding(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	var responseBody model.CreateRotateFetchBindingResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if settings.HeaderSettings.BrokerVersion > "2.15" && settings.BindingSettings.BindingMetadataSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected metadata to be not nil")
	}
	if (settings.HeaderSettings.BrokerVersion < "2.16" || !settings.BindingSettings.BindingMetadataSettings.ReturnMetadata) && responseBody.Metadata != nil {
		t.Errorf("Expected metadata to be nil")
	}
	if settings.HeaderSettings.BrokerVersion > "2.14" && settings.BindingSettings.BindingEndpointSettings.ReturnEndpoints && responseBody.Endpoints == nil {
		t.Errorf("Expected endpoints to be not nil")
	}
	if (settings.HeaderSettings.BrokerVersion < "2.15" || !settings.BindingSettings.BindingEndpointSettings.ReturnEndpoints) && responseBody.Endpoints != nil {
		t.Errorf("Expected endpoints to be nil")
	}
	if settings.BindingSettings.BindingVolumeMountSettings.ReturnVolumeMounts && firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "volume_mount") < len(firstOffering.Requires) && responseBody.VolumeMounts == nil {
		t.Errorf("Expected volumeMounts to be not nil")
	}
	if (!settings.BindingSettings.BindingVolumeMountSettings.ReturnVolumeMounts || firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "volume_mount") >= len(firstOffering.Requires)) && responseBody.VolumeMounts != nil {
		t.Errorf("Expected volumeMounts to be nil")
	}
	if settings.BindingSettings.ReturnRouteServiceURL && responseBody.RouteServiceUrl == nil {
		t.Errorf("Expected routeServiceURL to be not nil")
	}
	if !settings.BindingSettings.ReturnRouteServiceURL && responseBody.RouteServiceUrl != nil {
		t.Errorf("Expected routeServiceURL to be nil")
	}
	if settings.BindingSettings.ReturnSyslogDrainURL && firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "syslog_drain") < len(firstOffering.Requires) && responseBody.SyslogDrainUrl == nil {
		t.Errorf("Expected syslogDrainURL to be not nil")
	}
	if (!settings.BindingSettings.ReturnSyslogDrainURL || firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "syslog_drain") >= len(firstOffering.Requires)) && responseBody.SyslogDrainUrl != nil {
		log.Println(*responseBody.SyslogDrainUrl)
		t.Errorf("Expected syslogDrainURL to be nil")
	}
	if settings.BindingSettings.ReturnCredentials && firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "route_forwarding") < len(firstOffering.Requires) && responseBody.Credentials == nil {
		t.Errorf("Expected credentials to be not nil")
	}
	if (!settings.BindingSettings.ReturnCredentials || firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "route_forwarding") >= len(firstOffering.Requires)) && responseBody.Credentials != nil {
		t.Errorf("Expected credentials to be nil")
	}
	async := true
	requestSettings := model.RequestSettings{
		AsyncEndpoint: &async,
	}
	configBrokerSettings := struct {
		ReqSettings model.RequestSettings `json:"config_broker_settings"`
	}{
		ReqSettings: requestSettings,
	}
	bindingRequestBody = model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: configBrokerSettings,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingB+"?accepts_incomplete=true", requestBytes)
	if w.Code != 202 {
		log.Println(w.Body.String())
		t.Errorf("Expected 202, got %v", w.Code)
	}
	responseBody = model.CreateRotateFetchBindingResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if responseBody.Parameters != nil {
		t.Errorf("Expected parameters to be nil")
	}
	if responseBody.Metadata != nil {
		t.Errorf("Expected metadata to be nil")
	}
	if responseBody.Endpoints != nil {
		t.Errorf("Expected endpoints to be nil")
	}
	if responseBody.VolumeMounts != nil {
		t.Errorf("Expected volumeMounts to be nil")
	}
	if responseBody.RouteServiceUrl != nil {
		t.Errorf("Expected routeServiceURL to be nil")
	}
	if responseBody.SyslogDrainUrl != nil {
		t.Errorf("Expected syslogDrainURL to be nil")
	}
	if responseBody.Credentials != nil {
		t.Errorf("Expected credentials to be nil")
	}
	if settings.BindingSettings.ReturnOperationIfAsync && responseBody.Operation == nil {
		t.Errorf("Expected operation to be not nil")
	}
	if !settings.BindingSettings.ReturnOperationIfAsync && responseBody.Operation != nil {

	}
}

func TestCreateBindingIDInUse(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}

	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody = model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 409 {
		t.Errorf("Expected StatusCode 409, got %v", w.Code)
	}
}

func TestCreateBindingIdentical(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
}

func TestCreateBindingMissingInstance(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	requestBytes := new(bytes.Buffer)
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
}

//Check for correct behaviour when passing wrong service/plan id or omitting a required field
func TestCreateBindingInvalid(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	wrongService := "wrongService"
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &wrongService,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	wrongPlan := "wrongPlan"
	bindingRequestBody = model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &wrongPlan,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	bindingRequestBody = model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     nil,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
}

//Check for correct behaviour when fetching a binding
func TestFetchBinding(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	if firstOffering.BindingsRetrievable == nil || (firstOffering.BindingsRetrievable != nil && !*firstOffering.BindingsRetrievable) {
		t.Errorf("Test not applicable because the binding is not retrievable")
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID, nil)
	if w.Code != 200 {
		t.Errorf("Exoected StatusCode 200, got %v", w.Code)
	}
	responseBody := model.CreateRotateFetchBindingResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if settings.HeaderSettings.BrokerVersion > "2.15" && settings.BindingSettings.BindingMetadataSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected metadata to be not nil")
	}

	if (settings.HeaderSettings.BrokerVersion < "2.16" || !settings.BindingSettings.BindingMetadataSettings.ReturnMetadata) && responseBody.Metadata != nil {
		t.Errorf("Expected metadata to be nil")
	}
	if settings.BindingSettings.ReturnCredentials && responseBody.Credentials == nil {
		t.Errorf("Expected credentials to be not nil")
	}

	if !settings.BindingSettings.ReturnCredentials && responseBody.Credentials != nil {
		t.Errorf("Expected credentials to be nil")
	}
	if settings.BindingSettings.ReturnSyslogDrainURL && firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "syslog_drain") < len(firstOffering.Requires) && responseBody.SyslogDrainUrl == nil {
		t.Errorf("Expected syslogDrainURL to be not nil")
	}

	if (!settings.BindingSettings.ReturnSyslogDrainURL || firstOffering.Requires == nil || firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "syslog_drain") >= len(firstOffering.Requires)) && responseBody.SyslogDrainUrl != nil {
		t.Errorf("Expected syslogDrainURL to be nil")
	}
	if settings.BindingSettings.ReturnRouteServiceURL && firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "route_forwarding") < len(firstOffering.Requires) && responseBody.RouteServiceUrl == nil {
		t.Errorf("Expected routeServiceURL to be not nil")
	}

	if (!settings.BindingSettings.ReturnRouteServiceURL || firstOffering.Requires == nil || firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "route_forwarding") >= len(firstOffering.Requires)) && responseBody.RouteServiceUrl != nil {
		t.Errorf("Expected routeServiceURL to be nil")
	}
	if settings.BindingSettings.BindingVolumeMountSettings.ReturnVolumeMounts && firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "volume_mount") < len(firstOffering.Requires) && responseBody.VolumeMounts == nil {
		t.Errorf("Expected volumeMounts to be not nil")
	}

	if (!settings.BindingSettings.BindingVolumeMountSettings.ReturnVolumeMounts || firstOffering.Requires == nil || firstOffering.Requires != nil &&
		sort.SearchStrings(firstOffering.Requires, "volume_mount") >= len(firstOffering.Requires)) && responseBody.VolumeMounts != nil {
		t.Errorf("Expected volumeMounts to be nil")
	}
	if settings.BindingSettings.ReturnParameters && responseBody.Parameters == nil {
		t.Errorf("Expected parameters to be not nil")
	}

	if !settings.BindingSettings.ReturnParameters && responseBody.Parameters != nil {
		t.Errorf("Expected parameters to be nil")
	}
	if settings.HeaderSettings.BrokerVersion > "2.14" && settings.BindingSettings.BindingEndpointSettings.ReturnEndpoints && responseBody.Endpoints == nil {
		t.Errorf("Expected endpoints to be not nil")
	}

	if !settings.BindingSettings.BindingEndpointSettings.ReturnEndpoints && responseBody.Endpoints != nil {
		t.Errorf("Expected endpoints to be nil")
	}
}

//Check for correct behaviour when passing wrong, service/plan id
func TestFetchBindingInvalid(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	if firstOffering.BindingsRetrievable == nil || (firstOffering.BindingsRetrievable != nil && !*firstOffering.BindingsRetrievable) {
		t.Errorf("Test not applicable because the binding is not retrievable")
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/wrongBinding", nil)
	if w.Code != 404 {
		t.Errorf("Exoected StatusCode 404, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id=wrongService", nil)
	if w.Code != 400 {
		t.Errorf("Exoected StatusCode 400, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?plan_id=wrongPlan", nil)
	if w.Code != 400 {
		t.Errorf("Exoected StatusCode 400, got %v", w.Code)
	}
}

//Check for correct behaviour when passing wrong instance/binding does not exist
func TestFetchBindingMissing(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	w := performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/wrongBinding", nil)
	if w.Code != 404 {
		t.Errorf("Exoected StatusCode 404, got %v", w.Code)
	}
}

//Check for correct behaviour when unbinding (sync/async)
func TestUnbind(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.Unbind)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID+"&plan_id="+firstPlan.ID, nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingB+"?accepts_incomplete=true", requestBytes)
	if w.Code != 201 {
		log.Println(w.Body.String())
		t.Errorf("Expected 201, got %v", w.Code)
	}
	async := true
	requestSettings := model.RequestSettings{
		AsyncEndpoint: &async,
	}
	configBrokerSettings := struct {
		ReqSettings model.RequestSettings `json:"config_broker_settings"`
	}{
		ReqSettings: requestSettings,
	}
	deleteRequestBody := model.DeleteRequest{Parameters: configBrokerSettings}
	err = json.NewEncoder(requestBytes).Encode(deleteRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID+"&plan_id="+firstPlan.ID+"&accepts_incomplete=true", requestBytes)
	if w.Code != 202 {
		t.Errorf("Expected StatusCode 202, got %v", w.Code)
	}
	responseBody := model.OperationResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if settings.BindingSettings.ReturnOperationIfAsync && responseBody.Operation == nil {
		t.Errorf("Expected operation to be not nil")
	}
	if !settings.BindingSettings.ReturnOperationIfAsync && responseBody.Operation != nil {
		t.Errorf("Expected operation to be nil")
	}
}

//Check for correct behaviour when binding does not exist
func TestUnbindMissing(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.Unbind)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	w := performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID+"&plan_id="+firstPlan.ID, nil)
	if w.Code != 410 {
		t.Errorf("Expected StatusCode 410, got %v", w.Code)
	}
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceB, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceB+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID+"&plan_id="+firstPlan.ID, nil)
	if w.Code != 410 {
		t.Errorf("Expected StatusCode 410, got %v", w.Code)
	}
}

//Check for correct behaviour when passing wrong service/plan id or omitting them
func TestUnbindInvalid(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.Unbind)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id=wrongService&plan_id="+firstPlan.ID, nil)
	if w.Code != 412 {
		t.Errorf("Expected StatusCode 412, got %v", w.Code)
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID, nil)
	if w.Code != 412 {
		t.Errorf("Expected StatusCode 412, got %v", w.Code)
	}
}

//Check for correct behaviour when fetching a binding that has been deleted
func TestTestFetchBindingDeleted(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.Unbind)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{
		A: "I am field A",
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Could not encode request body")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID+"&plan_id="+firstPlan.ID, nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	if firstOffering.BindingsRetrievable == nil || (firstOffering.BindingsRetrievable != nil && !*firstOffering.BindingsRetrievable) {
		t.Errorf("Test not applicable because the binding is not retrievable")
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID, nil)
	if w.Code != 404 {
		t.Errorf("Exoected StatusCode 404, got %v", w.Code)
	}
}

func TestPollLastOperationBinding(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", bindingController.PollOperationState)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	async := true
	requestSettings := model.RequestSettings{
		AsyncEndpoint: &async,
	}
	configBrokerSettings := struct {
		ReqSettings model.RequestSettings `json:"config_broker_settings"`
	}{
		ReqSettings: requestSettings,
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: configBrokerSettings,
	}

	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?accepts_incomplete=true", requestBytes)
	if w.Code != 202 {
		t.Errorf("Expected StatusCode 202, got %v", w.Code)
	}
	bindingCreationResponseBody := model.CreateRotateFetchBindingResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &bindingCreationResponseBody)
	if bindingCreationResponseBody.Operation != nil {
		w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"/last_operation?operation="+*bindingCreationResponseBody.Operation, nil)
	} else {
		w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"/last_operation", nil)
	}
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	responseBody := model.BindingOperationPollResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if responseBody.State != "in progress" && responseBody.State != "succeeded" && responseBody.State != "failed" {
		t.Errorf("Expected statse to be either \"in progess\", \"succeeded\" or \"failed\", got %v", responseBody.State)
	}
	if settings.BindingSettings.ReturnDescriptionLastOperation && responseBody.Description == nil {
		t.Errorf("Expected description to be not nil")
	}
	if !settings.BindingSettings.ReturnDescriptionLastOperation && responseBody.Description != nil {
		t.Errorf("Expected description to be nil")
	}
}

func TestPollLastOperationBindingMissing(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", bindingController.PollOperationState)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"/last_operation", nil)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
}

func TestPollLastOperationBindingInvalid(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", bindingController.PollOperationState)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}

	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?accepts_incomplete=true", requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/wrongBinding/last_operation", nil)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"/last_operation?service_id=wrongService", nil)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"/last_operation?service_id="+firstOffering.ID+"&plan_id=wrongPlan", nil)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
}

func TestPollLastOperationBindingDeleted(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation", bindingController.PollOperationState)
	router.DELETE("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.Unbind)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}

	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?accepts_incomplete=true", requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}

	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID+"&plan_id="+firstPlan.ID, nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"/last_operation", nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	responseBody := model.BindingOperationPollResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Could not unmarshal response body to struct")
	}
	if responseBody.State != "failed" {
		t.Errorf("Expected state=failed, got %v", responseBody.State)
	}

	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	async := true
	requestSettings := model.RequestSettings{
		AsyncEndpoint: &async,
	}
	configBrokerSettings := struct {
		ReqSettings model.RequestSettings `json:"config_broker_settings"`
	}{
		ReqSettings: requestSettings,
	}
	deleteRequestBody := model.DeleteRequest{Parameters: configBrokerSettings}
	err = json.NewEncoder(requestBytes).Encode(deleteRequestBody)
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID+"&plan_id="+firstPlan.ID+"&accepts_incomplete=true", requestBytes)
	if w.Code != 202 {
		t.Errorf("Expected StatusCode 202, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"/last_operation", nil)
	if w.Code != 410 {
		t.Errorf("Expected StatusCode 4100, got %v", w.Code)
	}
	responseBody = model.BindingOperationPollResponse{}
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Could not unmarshal response body to struct")
	}
	if responseBody.State != "succeeded" {
		t.Errorf("Expected state=suceeded, got %v", responseBody.State)
	}
}

func TestBindingRotation(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
		return
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
		return
	}
	if settings.HeaderSettings.BrokerVersion < "2.17" {
		log.Println("Test not applicable since the given broker version does not support binding rotation")
		return
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingID := BindingA
	rotateRequestBody := model.RotateBindingRequest{
		PredecessorBindingId: &bindingID,
	}
	err = json.NewEncoder(requestBytes).Encode(rotateRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingB, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	if firstOffering.BindingsRetrievable == nil || (firstOffering.BindingsRetrievable != nil && !*firstOffering.BindingsRetrievable) {
		t.Errorf("Test not applicable because the binding is not retrievable")
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingB+"?service_id="+firstOffering.ID, nil)
	if w.Code != 200 {
		t.Errorf("Exoected StatusCode 200, got %v", w.Code)
	}
}

func TestBindingRotationIDInUse(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
		return
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
		return
	}
	if settings.HeaderSettings.BrokerVersion < "2.17" {
		log.Println("Test not applicable since the given broker version does not support binding rotation")
		return
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	parameters := struct {
		A string `json:"a"`
	}{A: "randomParameter"}
	bindingRequestBody = model.CreateBindingRequest{
		ServiceID:  &firstOffering.ID,
		PlanID:     &firstPlan.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingB, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingID := BindingA
	rotateRequestBody := model.RotateBindingRequest{
		PredecessorBindingId: &bindingID,
	}
	err = json.NewEncoder(requestBytes).Encode(rotateRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingB, requestBytes)
	if w.Code != 409 {
		t.Errorf("Expected StatusCode 409, got %v", w.Code)
	}
}

func TestBindingRotationIdentical(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
		return
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
		return
	}
	if settings.HeaderSettings.BrokerVersion < "2.17" {
		log.Println("Test not applicable since the given broker version does not support binding rotation")
		return
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingID := BindingA
	rotateRequestBody := model.RotateBindingRequest{
		PredecessorBindingId: &bindingID,
	}
	err = json.NewEncoder(requestBytes).Encode(rotateRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if !settings.BindingSettings.StatusCodeOKPossible {
		t.Errorf("Test not applicable since returning 200 for identical bindings is set to false")
	} else {
		if w.Code != 200 {
			t.Errorf("Expected StatusCode 200, got %v", w.Code)
		}
	}
}

func TestBindingRotationMissing(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
		return
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
		return
	}
	if settings.HeaderSettings.BrokerVersion < "2.17" {
		log.Println("Test not applicable since the given broker version does not support binding rotation")
		return
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingID := "noBinding"
	rotateRequestBody := model.RotateBindingRequest{
		PredecessorBindingId: &bindingID,
	}
	err = json.NewEncoder(requestBytes).Encode(rotateRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingB, requestBytes)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
	bindingID = BindingA
	err = json.NewEncoder(requestBytes).Encode(rotateRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/wrongInstance/service_bindings/"+BindingB, requestBytes)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
}

//Check for correct behaviour when rotating a binding but giving a different existing instance_id for the new binding
//and using the binding of the other instance as predecessor
//Expecting 404, since the program will look for bindings of the given instance_id and the predecessor is a binding of
//other instance (the one that is not given) which will not be found
func TestBindingRotationInvalid(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
		return
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
		return
	}
	if settings.HeaderSettings.BrokerVersion < "2.17" {
		log.Println("Test not applicable since the given broker version does not support binding rotation")
		return
	}
	var serviceInstances map[string]*model.ServiceDeployment
	serviceInstances = make(map[string]*model.ServiceDeployment)
	var platform string
	deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
	deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
	var bindingInstances map[string]*model.ServiceBinding
	bindingInstances = make(map[string]*model.ServiceBinding)
	bindingService := service.NewBindingService(&serviceInstances, &bindingInstances, settings, catalog)
	bindingController := controller.NewBindingController(bindingService, settings, &platform)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PUT("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.CreateBinding)
	router.GET("/v2/service_instances/:instance_id/service_bindings/:binding_id", bindingController.FetchBinding)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	requestBytes := new(bytes.Buffer)
	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}

	err = json.NewEncoder(requestBytes).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceB, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingRequestBody := model.CreateBindingRequest{
		ServiceID: &firstOffering.ID,
		PlanID:    &firstPlan.ID,
	}
	err = json.NewEncoder(requestBytes).Encode(bindingRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA, requestBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	bindingID := BindingA
	rotateRequestBody := model.RotateBindingRequest{
		PredecessorBindingId: &bindingID,
	}
	err = json.NewEncoder(requestBytes).Encode(rotateRequestBody)
	if err != nil {
		t.Errorf("Error while encoding the request body to bytes")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceB+"/service_bindings/"+BindingB, requestBytes)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
}
