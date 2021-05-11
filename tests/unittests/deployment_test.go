package unittests

import (
	"github.com/evoila/configurable-test-osb/model"
	"testing"
)

func TestNewServiceDeploymentReturnsCorrectValues(t *testing.T) {
	instanceID := "IAmInstance"
	settings := model.Settings{
		ProvisionSettings: model.ProvisionSettings{
			CreateDashboardURL: false,
			CreateMetadata:     false,
		},
	}
	plan := model.ServicePlan{
		ID: "PlanOne",
	}
	instancesRetrievable := false
	planSlice := []*model.ServicePlan{&plan}
	offering := model.ServiceOffering{
		ID:                   "OfferingOne",
		Plans:                &planSlice,
		InstancesRetrievable: &instancesRetrievable,
	}

	orgGUID := "TestOrg"
	spaceGUID := "SpaceGUID"
	offeringSlice := []*model.ServiceOffering{&offering}
	catalog := model.Catalog{ServiceOfferings: &offeringSlice}
	request := model.ProvideServiceInstanceRequest{
		ServiceID:        &offering.ID,
		PlanID:           &plan.ID,
		OrganizationGUID: &orgGUID,
		SpaceGUID:        &spaceGUID,
	}
	deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
	if deployment == nil {
		t.Fatalf("Expected object for deployment, got %v", deployment)
	}
	if deployment.DashboardURL() != nil {
		t.Errorf("Expected deployment.DashboardURL() == nil but got %v", deployment.DashboardURL())
	}
	if deployment.Metadata() != nil {
		t.Errorf("Expected deployment.Metadata() == nil but got %v", deployment.Metadata())
	}
	if deployment.FetchResponse() != nil {
		t.Errorf("Expected deployment.FetchResponse() == nil but got %v", deployment.FetchResponse())
	}
}

func TestNewServiceDeploymentWithDashboardURL(t *testing.T) {
	instanceID := "IAmInstance"
	settings := model.Settings{
		ProvisionSettings: model.ProvisionSettings{
			CreateDashboardURL: true,
			CreateMetadata:     false,
		},
	}
	plan := model.ServicePlan{
		ID: "PlanOne",
	}
	instancesRetrievable := false
	planSlice := []*model.ServicePlan{&plan}
	offering := model.ServiceOffering{
		ID:                   "OfferingOne",
		Plans:                &planSlice,
		InstancesRetrievable: &instancesRetrievable,
	}
	orgGUID := "TestOrg"
	spaceGUID := "SpaceGUID"
	offeringSlice := []*model.ServiceOffering{&offering}
	catalog := model.Catalog{ServiceOfferings: &offeringSlice}
	request := model.ProvideServiceInstanceRequest{
		ServiceID:        &offering.ID,
		PlanID:           &plan.ID,
		OrganizationGUID: &orgGUID,
		SpaceGUID:        &spaceGUID,
	}
	deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
	if deployment == nil {
		t.Fatalf("Expected object for deployment, got %v", deployment)
	}
	if deployment.DashboardURL() == nil {
		t.Errorf("Expected value for deployment.DashboardURL(), but got nil")
	}
}

func TestNewServiceDeploymentWithMetadata(t *testing.T) {
	instanceID := "IAmInstance"
	settings := model.Settings{
		ProvisionSettings: model.ProvisionSettings{
			CreateDashboardURL: false,
			CreateMetadata:     true,
		},
	}
	plan := model.ServicePlan{
		ID: "PlanOne",
	}
	instancesRetrievable := false
	planSlice := []*model.ServicePlan{&plan}
	offering := model.ServiceOffering{
		ID:                   "OfferingOne",
		Plans:                &planSlice,
		InstancesRetrievable: &instancesRetrievable,
	}
	orgGUID := "TestOrg"
	spaceGUID := "SpaceGUID"
	offeringSlice := []*model.ServiceOffering{&offering}
	catalog := model.Catalog{ServiceOfferings: &offeringSlice}
	request := model.ProvideServiceInstanceRequest{
		ServiceID:        &offering.ID,
		PlanID:           &plan.ID,
		OrganizationGUID: &orgGUID,
		SpaceGUID:        &spaceGUID,
	}
	deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
	if deployment == nil {
		t.Fatalf("Expected object for deployment, got %v", deployment)
	}
	if deployment.Metadata() == nil {
		t.Errorf("Expected value for deployment.Metadata(), but got nil")
	}
}

func TestNewServiceDeploymentIsRetrievable(t *testing.T) {
	instanceID := "IAmInstance"
	settings := model.Settings{
		ProvisionSettings: model.ProvisionSettings{
			CreateDashboardURL: false,
			CreateMetadata:     false,
		},
	}
	plan := model.ServicePlan{
		ID: "PlanOne",
	}
	instancesRetrievable := true
	planSlice := []*model.ServicePlan{&plan}
	offering := model.ServiceOffering{
		ID:                   "OfferingOne",
		Plans:                &planSlice,
		InstancesRetrievable: &instancesRetrievable,
	}
	orgGUID := "TestOrg"
	spaceGUID := "SpaceGUID"
	offeringSlice := []*model.ServiceOffering{&offering}
	catalog := model.Catalog{ServiceOfferings: &offeringSlice}
	request := model.ProvideServiceInstanceRequest{
		ServiceID:        &offering.ID,
		PlanID:           &plan.ID,
		OrganizationGUID: &orgGUID,
		SpaceGUID:        &spaceGUID,
	}
	deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
	if deployment == nil {
		t.Fatalf("Expected object for deployment, got %v", deployment)
	}
	if deployment.FetchResponse() == nil {
		t.Errorf("Expected value for deployment.FetchResponse(), but got nil")
	}
}
