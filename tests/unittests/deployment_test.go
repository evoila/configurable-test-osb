package unittests

import (
	"github.com/evoila/configurable-test-osb/model"
	"testing"
)

func TestNewServiceDeployment(t *testing.T) {
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
	t.Run("TestNewServiceDeploymentReturnsCorrectValues", func(t *testing.T) {
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
	})
	settings.ProvisionSettings.CreateDashboardURL = true
	t.Run("TestNewServiceDeploymentWithDashboardURL", func(t *testing.T) {
		deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
		if deployment == nil {
			t.Fatalf("Expected object for deployment, got %v", deployment)
		}
		if deployment.DashboardURL() == nil {
			t.Errorf("Expected value for deployment.DashboardURL(), but got nil")
		}
	})
	settings.ProvisionSettings.CreateDashboardURL = false
	settings.ProvisionSettings.CreateMetadata = true
	t.Run("TestNewServiceDeploymentWithMetadata", func(t *testing.T) {
		deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
		if deployment == nil {
			t.Fatalf("Expected object for deployment, got %v", deployment)
		}
		if deployment.Metadata() == nil {
			t.Errorf("Expected value for deployment.Metadata(), but got nil")
		}
	})
	settings.ProvisionSettings.CreateMetadata = false
	instancesRetrievable = true
	t.Run("TestNewServiceDeploymentIsRetrievable", func(t *testing.T) {
		deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
		if deployment == nil {
			t.Fatalf("Expected object for deployment, got %v", deployment)
		}
		if deployment.FetchResponse() == nil {
			t.Errorf("Expected value for deployment.FetchResponse(), but got nil")
		}
	})
}

func TestUpdate(t *testing.T) {
	instanceID := "IAmInstance"
	settings := model.Settings{
		ProvisionSettings: model.ProvisionSettings{
			CreateDashboardURL: false,
			CreateMetadata:     false,
		},
	}
	planOne := model.ServicePlan{
		ID: "PlanOne",
	}
	planTwo := model.ServicePlan{
		ID: "PlanTwo",
	}
	instancesRetrievable := false
	planSlice := []*model.ServicePlan{&planOne, &planTwo}
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
		PlanID:           &planOne.ID,
		OrganizationGUID: &orgGUID,
		SpaceGUID:        &spaceGUID,
	}
	deployment, _ := model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
	if deployment == nil {
		t.Fatalf("Expected object for deployment, got %v", deployment)
	}
	updateServiceInstanceRequest := model.UpdateServiceInstanceRequest{
		ServiceId: &offering.ID,
	}

	t.Run("TestUpdateReturnsCorrectValues", func(t *testing.T) {
		operationID := deployment.Update(&updateServiceInstanceRequest)
		if operationID == nil {
			t.Errorf("Expected value for deployment.Update(*ServiceDeployment), but got nil")
		}
	})
	planUpdateable := true
	offering.PlanUpdateable = &planUpdateable
	updateServiceInstanceRequest.PlanId = &planTwo.ID
	t.Run("TestUpdatePlanApplied", func(t *testing.T) {
		operationID := deployment.Update(&updateServiceInstanceRequest)
		if operationID == nil {
			t.Errorf("Expected value for deployment.Update(*ServiceDeployment), but got nil")
		}
		if *deployment.PlanID() != planTwo.ID {
			t.Errorf("Expected *deployment.PlanID() == %v, but got %v", planTwo.ID, *deployment.PlanID())
			t.Log(*updateServiceInstanceRequest.PlanId)
		}
	})
	updateServiceInstanceRequest.PlanId = nil
}
