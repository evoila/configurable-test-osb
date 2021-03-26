package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/server"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"
)

const InstanceA = "instanceOne"
const InstanceB = "instanceTwo"
const BindingA = "bindingOne"
const BindingB = "bindingTwo"

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

//Tests, if catalog is valid
func TestCatalog(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	catalogService := service.NewCatalogService(catalog)
	catalogController := controller.NewCatalogController(&catalogService, settings)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/v2/catalog", catalogController.GetCatalog)
	w := performRequest(router, "GET", "/v2/catalog", nil)
	catalogJson, err := os.Open("config/catalog.json")
	if err != nil {
		t.Errorf("Could not open config/catalog.json")
	}
	byteVal, err := ioutil.ReadAll(catalogJson)
	require.JSONEq(t, w.Body.String(), string(byteVal))
}

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
	if settings.ProvisionSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected the field metadata to be not empty")
	}
	if !settings.ProvisionSettings.ReturnMetadata && responseBody.Metadata != nil {
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
	if settings.BindingSettings.BindingMetadataSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected metadata to be not nil")
	}
	if !settings.BindingSettings.BindingMetadataSettings.ReturnMetadata && responseBody.Metadata != nil {
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
	if settings.BindingSettings.BindingMetadataSettings.ReturnMetadata && responseBody.Metadata == nil {
		t.Errorf("Expected metadata to be not nil")
	}

	if !settings.BindingSettings.BindingMetadataSettings.ReturnMetadata && responseBody.Metadata != nil {
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
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
	}
	w = performRequest(router, "DELETE", "/v2/service_instances/"+InstanceA+"/service_bindings/"+BindingA+"?service_id="+firstOffering.ID, nil)
	if w.Code != 400 {
		t.Errorf("Expected StatusCode 400, got %v", w.Code)
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
