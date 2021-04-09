package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/MaxFuhrich/configurable-test-osb/controller"
	"github.com/MaxFuhrich/configurable-test-osb/model"
	"github.com/MaxFuhrich/configurable-test-osb/server"
	"github.com/MaxFuhrich/configurable-test-osb/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

//Test, if a valid provision request works as intended
func TestProvision(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	reqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	if w.Code != 201 {
		t.Errorf("Expected StatusCode 201, got %v", w.Code)
	}
	var responseBody model.ProvideUpdateServiceInstanceResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Could not unmarshal response body to response struct")
	}
	if settings.ProvisionSettings.ReturnDashboardURL && responseBody.DashboardUrl == nil {
		t.Errorf("Expected the field dashboardURL in the response to be not empty")
	}
	if !settings.ProvisionSettings.ReturnDashboardURL && responseBody.DashboardUrl != nil {
		t.Errorf("Got dashboardURL which should not be returned")
	}
	if settings.HeaderSettings.BrokerVersion > "2.15" && settings.ProvisionSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected the field metadata to be not empty")
	}
	if (settings.HeaderSettings.BrokerVersion < "2.16" || !settings.ProvisionSettings.ReturnMetadata) && responseBody.Metadata != nil {
		t.Errorf("Got metadata which should not be returned")
	}
	if responseBody.Operation != nil {
		t.Errorf("Got operation which should not be returned")
	}
}

//Tests for correct behaviour if instance_id already exists
func TestProvisionDuplicate(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
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
	provisionA := new(bytes.Buffer)
	err = json.NewEncoder(provisionA).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionA)
	organizationGUIDB := "different organization"
	spaceGUIDB := "space"
	requestBodyB := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUIDB,
		SpaceGUID:        &spaceGUIDB,
	}
	provisionB := new(bytes.Buffer)
	err = json.NewEncoder(provisionB).Encode(requestBodyB)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionB)
	if w.Code != 409 {
		t.Errorf("Expected StatusCode 409, got %v", w.Code)
	}
}

//Tests for correct behaviour if instance_id already exists and requested and existing deployment are identical
func TestProvisionDuplicateIdentical(t *testing.T) {
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
		return
	}
	if settings == nil {
		t.Errorf("Settings were not assigned!")
	}
	if settings.ProvisionSettings.StatusCodeOKPossibleForIdenticalProvision {
		catalog, err := server.MakeCatalog()
		if err != nil || catalog == nil {
			t.Errorf("Catalog could not be created!")
			return
		}
		var serviceInstances map[string]*model.ServiceDeployment
		serviceInstances = make(map[string]*model.ServiceDeployment)
		var platform string
		deploymentService := service.NewDeploymentService(catalog, &serviceInstances, settings)
		deploymentController := controller.NewDeploymentController(deploymentService, settings, &platform)
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		router.Use(gin.Recovery())
		router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
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
		provisionA := new(bytes.Buffer)
		err = json.NewEncoder(provisionA).Encode(requestBodyA)
		if err != nil {
			t.Errorf("Could not create []byte from struct!")
		}
		performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionA)
		err = json.NewEncoder(provisionA).Encode(requestBodyA)
		if err != nil {
			t.Errorf("Could not create []byte from struct!")
		}
		w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionA)
		if w.Code != 200 {
			t.Errorf("Expected StatusCode 200, got %v", w.Code)
		}
	} else {
		log.Println("Test not applicable, because the settings forbid sending status code 200 when provisioning.")
	}
}

//Tests for correct behaviour, if provision request is invalid (wrong service and plan id)
func TestProvisionInvalid(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	serviceID := "this offering does not exist"
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &serviceID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}

	provisionA := new(bytes.Buffer)
	err = json.NewEncoder(provisionA).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionA)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	planID := "this plan does not exist"
	organizationGUIDB := "organization"
	spaceGUIDB := "space"
	requestBodyB := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &planID,
		OrganizationGUID: &organizationGUIDB,
		SpaceGUID:        &spaceGUIDB,
	}
	provisionB := new(bytes.Buffer)
	err = json.NewEncoder(provisionB).Encode(requestBodyB)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionB)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
}

//Test for correct behaviour when fetching existing instance
func TestFetchInstanceValid(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.GET("/v2/service_instances/:instance_id", deploymentController.FetchServiceInstance)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	parameters := struct {
		TestWord string
	}{TestWord: "Hello"}

	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
		Parameters:       parameters,
	}
	provisionA := new(bytes.Buffer)
	err = json.NewEncoder(provisionA).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionA)
	w := performRequest(router, "GET", "/v2/service_instances/"+InstanceA, nil)
	if firstOffering.InstancesRetrievable != nil && !*firstOffering.InstancesRetrievable && w.Code != 400 {
		t.Errorf("Instance of the offering should not be fetchable. Expected StatusCode 400, got %v", w.Code)
	}
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	var responseBody model.FetchingServiceInstanceResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if settings.FetchServiceInstanceSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected metadata but field was empty")
	}
	if !settings.FetchServiceInstanceSettings.ReturnMetadata && responseBody.Metadata != nil {
		t.Errorf("Got metadata which should not be returned")
	}
	if settings.FetchServiceInstanceSettings.ReturnDashboardURL && responseBody.DashboardUrl == nil {
		t.Errorf("Expected dashboard url but field was empty")
	}
	if !settings.FetchServiceInstanceSettings.ReturnDashboardURL && responseBody.DashboardUrl != nil {
		t.Errorf("Got dashboard url which should not be returned")
	}
	if settings.FetchServiceInstanceSettings.ReturnParameters && responseBody.Parameters == nil {
		t.Errorf("Expected parameters but field was empty")
	}
	if !settings.FetchServiceInstanceSettings.ReturnParameters && responseBody.Parameters != nil {
		t.Errorf("Got parameters which should not be returned")
	}
	if settings.FetchServiceInstanceSettings.ReturnMaintenanceInfo && responseBody.MaintenanceInfo == nil {
		t.Errorf("Expected maintenance info but field was empty")
	}
	if !settings.FetchServiceInstanceSettings.ReturnMaintenanceInfo && responseBody.MaintenanceInfo != nil {
		t.Errorf("Got maintenance info which should not be returned")
	}
	if settings.FetchServiceInstanceSettings.ReturnPlanID && responseBody.PlanId == nil {
		t.Errorf("Expected plan id but field was empty")
	}
	if !settings.FetchServiceInstanceSettings.ReturnPlanID && responseBody.PlanId != nil {
		t.Errorf("Got plan id which should not be returned")
	}
	if settings.FetchServiceInstanceSettings.ReturnServiceID && responseBody.ServiceId == nil {
		t.Errorf("Expected service id but field was empty")
	}
	if !settings.FetchServiceInstanceSettings.ReturnServiceID && responseBody.ServiceId != nil {
		t.Errorf("Got service id which should not be reuturned")
	}
}

//Test for correct behaviour when fetching non-existent instance
func TestFetchInstanceMissing(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/v2/service_instances/:instance_id", deploymentController.FetchServiceInstance)
	w := performRequest(router, "GET", "/v2/service_instances/"+InstanceA, nil)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
}

//Test for correct behaviour when fetching existing instance with invalid parameters (service id wrong)
func TestFetchInstanceInvalid(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.GET("/v2/service_instances/:instance_id", deploymentController.FetchServiceInstance)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	parameters := struct {
		TestWord string
	}{TestWord: "Hello"}

	organizationGUID := "organization"
	spaceGUID := "space"
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
		Parameters:       parameters,
	}
	provisionA := new(bytes.Buffer)
	err = json.NewEncoder(provisionA).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionA)
	w := performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"?service_id=wrongID", nil)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
}

//Test for correct behaviour for valid request
func TestDeprovisioning(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.DELETE("/v2/service_instances/:instance_id", deploymentController.Delete)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	reqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	parameters := struct {
		TestWord string
	}{TestWord: "Hello"}
	deleteRequestBody := model.DeleteRequest{Parameters: parameters}
	delReqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(delReqBodyBytes).Encode(deleteRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	resource := fmt.Sprintf("%v?service_id=%v&plan_id=%v", InstanceA, firstOffering.ID, firstPlan.ID)
	w := performRequest(router, "DELETE", "/v2/service_instances/"+resource, delReqBodyBytes)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	require.JSONEq(t, w.Body.String(), "{}")
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	async := true
	requestSettings := model.RequestSettings{
		AsyncEndpoint: &async,
	}
	configBrokerSettings := struct {
		ReqSettings model.RequestSettings `json:"config_broker_settings"`
	}{ReqSettings: requestSettings}
	deleteRequestBody = model.DeleteRequest{Parameters: configBrokerSettings}
	err = json.NewEncoder(delReqBodyBytes).Encode(deleteRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+resource+"&accepts_incomplete=true", delReqBodyBytes)
	if w.Code != 202 {
		log.Println(w.Body.String())
		t.Errorf("Expected 202, got %v", w.Code)
	}
}

//Test for correct behaviour when instance does not exist
func TestDeprovisioningInstanceMissing(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.DELETE("/v2/service_instances/:instance_id", deploymentController.Delete)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	resource := fmt.Sprintf("%v?service_id=%v&plan_id=%v", InstanceA, &firstOffering.ID, &firstPlan.ID)
	w := performRequest(router, "DELETE", "/v2/service_instances/"+resource, nil)
	if w.Code != 410 {
		t.Errorf("Expected StatusCode 410, got %v", w.Code)
	}
}

//Test for correct behaviour, when service/plan id wrong
func TestDeprovisioningInvalid(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.DELETE("/v2/service_instances/:instance_id", deploymentController.Delete)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	reqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	parameters := struct {
		TestWord string
	}{TestWord: "Hello"}
	deleteRequestBody := model.DeleteRequest{Parameters: parameters}
	delReqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(delReqBodyBytes).Encode(deleteRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	resource := fmt.Sprintf("%v?service_id=%v&plan_id=%v", InstanceA, "wrongOffering", &firstPlan.ID)
	w := performRequest(router, "DELETE", "/v2/service_instances/"+resource, delReqBodyBytes)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	resource = fmt.Sprintf("%v?service_id=%v&plan_id=%v", InstanceA, &firstOffering.ID, "wrongPlan")
	w = performRequest(router, "DELETE", "/v2/service_instances/"+resource, delReqBodyBytes)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
}

//Test delete before
//Test for correct behaviour when fetching deleted instance
func TestFetchInstanceDeleted(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.DELETE("/v2/service_instances/:instance_id", deploymentController.Delete)
	router.GET("/v2/service_instances/:instance_id", deploymentController.FetchServiceInstance)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	reqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	resource := fmt.Sprintf("%v?service_id=%v&plan_id=%v", InstanceA, firstOffering.ID, firstPlan.ID)
	w := performRequest(router, "DELETE", "/v2/service_instances/"+resource, nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA, nil)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 410, got %v", w.Code)
	}
}

func TestPollLastOperationInstance(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.GET("/v2/service_instances/:instance_id/last_operation", deploymentController.PollOperationState)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	reqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	w := performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/last_operation?service_id="+firstOffering.ID, nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	var responseBody model.InstanceOperationPollResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Could not unmarshal response body to response struct")
	}
	if responseBody.State != "succeeded" {
		t.Errorf("Expected \"state\": \"succeeded\", got %v", responseBody.State)
	}
	if settings.PollInstanceOperationSettings.DescriptionInResponse && responseBody.Description == nil {
		t.Errorf("Expected \"description\" to be not nil")
	}
	if !settings.PollInstanceOperationSettings.DescriptionInResponse && responseBody.Description != nil {
		t.Errorf("Expected \"description\" to be nil")
	}
}

func TestPollLastOperationInstanceMissing(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/v2/service_instances/:instance_id/last_operation", deploymentController.PollOperationState)
	w := performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/last_operation", nil)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
}

//Test for correct behaviour when passing wrong service/plan id
func TestPollLastOperationInstanceInvalid(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.GET("/v2/service_instances/:instance_id/last_operation", deploymentController.PollOperationState)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	reqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	w := performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/last_operation?service_id=wrongService", nil)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/last_operation?plan_id=wrongPlan", nil)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
}

//Test for correct behaviour when polling the last operation of a deleted instance
//If deleted async, it should return 410 (this will be the response for the deprovision)
//If deleted sync, it should return 200 with "state": "failed" (the response won't be for the deprovision, since it is
//created sync, but for the provision in case the state of the provisioning needs to be checked)
func TestPollLastOperationInstanceDeleted(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.DELETE("/v2/service_instances/:instance_id", deploymentController.Delete)
	router.GET("/v2/service_instances/:instance_id/last_operation", deploymentController.PollOperationState)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	firstOfferingPlans := firstOffering.Plans
	firstPlan := (*firstOfferingPlans)[0]
	organizationGUID := "organization"
	spaceGUID := "space"
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        &firstOffering.ID,
		PlanID:           &firstPlan.ID,
		OrganizationGUID: &organizationGUID,
		SpaceGUID:        &spaceGUID,
	}
	reqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	async := true
	requestSettings := model.RequestSettings{
		AsyncEndpoint: &async,
	}
	configBrokerSettings := struct {
		ReqSettings model.RequestSettings `json:"config_broker_settings"`
	}{ReqSettings: requestSettings}
	deleteRequestBody := model.DeleteRequest{Parameters: configBrokerSettings}
	delReqBodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(delReqBodyBytes).Encode(deleteRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	resource := fmt.Sprintf("%v?service_id=%v&plan_id=%v", InstanceA, firstOffering.ID, firstPlan.ID)
	w := performRequest(router, "DELETE", "/v2/service_instances/"+resource+"&accepts_incomplete=true", delReqBodyBytes)
	if w.Code != 202 {
		t.Errorf("Expected 202, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/last_operation?service_id="+firstOffering.ID, nil)
	if w.Code != 410 {
		t.Errorf("Expected StatusCode 410, got %v", w.Code)
	}
	err = json.NewEncoder(reqBodyBytes).Encode(requestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, reqBodyBytes)
	w = performRequest(router, "DELETE", "/v2/service_instances/"+resource+"&accepts_incomplete=true", nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	w = performRequest(router, "GET", "/v2/service_instances/"+InstanceA+"/last_operation?service_id="+firstOffering.ID, nil)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 410, got %v", w.Code)
	}
	var responseBody model.InstanceOperationPollResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Could not unmarshal response body to response struct")
	}
	if responseBody.State != "failed" {
		t.Errorf("Expected \"state\": \"failed\", got %v", responseBody.State)
	}
}

//Test for correct behaviour when sending an update request (sync/async)
func TestUpdate(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PATCH("/v2/service_instances/:instance_id", deploymentController.UpdateServiceInstance)
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
		TestWord string `json:"test_word"`
	}{TestWord: "Hello"}
	updateRequestBody := model.UpdateServiceInstanceRequest{
		ServiceId:  &firstOffering.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(updateRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	var responseBody model.ProvideUpdateServiceInstanceResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Could not unmarshal response body to struct")
	}
	if settings.ProvisionSettings.CreateDashboardURL && settings.ProvisionSettings.ReturnDashboardURL && responseBody.DashboardUrl == nil {
		t.Errorf("Expected dashboardURL to not be nil")
	}
	if !settings.ProvisionSettings.ReturnDashboardURL && responseBody.DashboardUrl != nil {
		t.Errorf("Expected dashboardURL to be nil")
	}
	if settings.ProvisionSettings.CreateMetadata && settings.ProvisionSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected metadata to not be nil")
	}
	if !settings.ProvisionSettings.ReturnMetadata && responseBody.Metadata != nil {
		t.Errorf("Expected metadata to be nil")
	}
	if responseBody.Operation != nil {
		t.Errorf("Expected operation to be nil")
	}
	async := true
	requestSettings := model.RequestSettings{
		AsyncEndpoint: &async,
	}
	configBrokerSettings := struct {
		ReqSettings model.RequestSettings `json:"config_broker_settings"`
		A           string                `json:"a"`
	}{
		ReqSettings: requestSettings,
		A:           "IAmA",
	}
	updateRequestBody = model.UpdateServiceInstanceRequest{
		ServiceId:  &firstOffering.ID,
		Parameters: configBrokerSettings,
	}
	err = json.NewEncoder(requestBytes).Encode(updateRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA+"?accepts_incomplete=true", requestBytes)
	if w.Code != 202 {
		log.Println(w.Body.String())
		t.Errorf("Expected StatusCode 202, got %v", w.Code)
	}
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Could not unmarshal response body to struct")
	}
	if settings.ProvisionSettings.ReturnOperationIfAsync && responseBody.Operation == nil {
		t.Errorf("Expected operation to not be nil")
	}
	if !settings.ProvisionSettings.ReturnOperationIfAsync && responseBody.Operation != nil {
		t.Errorf("Expected operation to be nil")
	}
}

//Test for correct behaviour, when nothing changed
func TestUpdateNoChanges(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PATCH("/v2/service_instances/:instance_id", deploymentController.UpdateServiceInstance)
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
		TestWord string `json:"test_word"`
	}{TestWord: "Hello"}
	updateRequestBody := model.UpdateServiceInstanceRequest{
		ServiceId:  &firstOffering.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(updateRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	err = json.NewEncoder(requestBytes).Encode(updateRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 200 {
		t.Errorf("Expected StatusCode 200, got %v", w.Code)
	}
	if settings.HeaderSettings.BrokerVersion > "2.14" {
		if w.Body.String() != "{}" {
			t.Errorf("Expected response body {}, got %v", w.Body.String())
		}
	}
}

//Check for correct behaviour when instance does not exist
func TestUpdateMissingInstance(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PATCH("/v2/service_instances/:instance_id", deploymentController.UpdateServiceInstance)
	offerings := catalog.ServiceOfferings
	firstOffering := (*offerings)[0]
	requestBytes := new(bytes.Buffer)

	parameters := struct {
		TestWord string `json:"test_word"`
	}{TestWord: "Hello"}
	updateRequestBody := model.UpdateServiceInstanceRequest{
		ServiceId:  &firstOffering.ID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(updateRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w := performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 404 {
		t.Errorf("Expected StatusCode 404, got %v", w.Code)
	}
}

//Check for correct behaviour when sending wrong service id, wrong previous plan id and omitting service id
func TestUpdateInvalid(t *testing.T) {
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.PUT("/v2/service_instances/:instance_id", deploymentController.Provision)
	router.PATCH("/v2/service_instances/:instance_id", deploymentController.UpdateServiceInstance)
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
	//wrong service id
	parameters := struct {
		TestWord string `json:"test_word"`
	}{TestWord: "Hello"}
	wrongServiceID := "nonExistentOffering"
	updateRequestBody := model.UpdateServiceInstanceRequest{
		ServiceId:  &wrongServiceID,
		Parameters: parameters,
	}
	err = json.NewEncoder(requestBytes).Encode(updateRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	//omitting service id
	w = performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA, nil)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	//wrong plan id in previous values
	wrongPlanID := "nonExistingPlan"
	previousValues := model.PreviousValues{
		PlanId: &wrongPlanID,
	}
	updateRequestBody = model.UpdateServiceInstanceRequest{
		ServiceId:      &firstOffering.ID,
		PreviousValues: &previousValues,
	}
	err = json.NewEncoder(requestBytes).Encode(updateRequestBody)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	w = performRequest(router, "PATCH", "/v2/service_instances/"+InstanceA, requestBytes)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
}
