package main

import (
	"bytes"
	"encoding/json"
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
	"testing"
)

const InstanceA = "instanceOne"

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
	requestBody := model.ProvideServiceInstanceRequest{
		ServiceID:        firstOffering.Id,
		PlanID:           firstPlan.ID,
		OrganizationGUID: "organization",
		SpaceGUID:        "space",
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
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        firstOffering.Id,
		PlanID:           firstPlan.ID,
		OrganizationGUID: "organization",
		SpaceGUID:        "space",
	}
	provisionA := new(bytes.Buffer)
	err = json.NewEncoder(provisionA).Encode(requestBodyA)
	if err != nil {
		t.Errorf("Could not create []byte from struct!")
	}
	performRequest(router, "PUT", "/v2/service_instances/"+InstanceA, provisionA)
	requestBodyB := model.ProvideServiceInstanceRequest{
		ServiceID:        firstOffering.Id,
		PlanID:           firstPlan.ID,
		OrganizationGUID: "different organization",
		SpaceGUID:        "space",
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
	if settings.ProvisionSettings.StatusCodeOKPossible {
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
		requestBodyA := model.ProvideServiceInstanceRequest{
			ServiceID:        firstOffering.Id,
			PlanID:           firstPlan.ID,
			OrganizationGUID: "organization",
			SpaceGUID:        "space",
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
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        "this offering does not exist",
		PlanID:           firstPlan.ID,
		OrganizationGUID: "organization",
		SpaceGUID:        "space",
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
	requestBodyB := model.ProvideServiceInstanceRequest{
		ServiceID:        firstOffering.Id,
		PlanID:           "this plan does not exist",
		OrganizationGUID: "organization",
		SpaceGUID:        "space",
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
	requestBodyA := model.ProvideServiceInstanceRequest{
		ServiceID:        firstOffering.Id,
		PlanID:           firstPlan.ID,
		OrganizationGUID: "organization",
		SpaceGUID:        "space",
	}
	//requestBodyA.Parameters =
	//word := "hello"
	//requestBodyA.Parameters = &word
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

}

//Test for correct behaviour when fetching existing instance with invalid parameters (service id wrong)
func TestFetchInstanceInvalid(t *testing.T) {

}

//Test delete before
//Test for correct behaviour when fetching deleted instance
func TestFetchInstanceDeleted(t *testing.T) {

}
