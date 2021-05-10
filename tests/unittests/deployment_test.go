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
	var deployment *model.ServiceDeployment
	var operationID *string
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				deployment, operationID = model.NewServiceDeployment(instanceID, &request, &settings, &catalog)
				if deployment == nil {
					t.Errorf("Expected deployment to not be nil, but got nil")
					return
				}
				if operationID == nil {
					t.Errorf("Expected operationID to not be nil, but got nil")
					return
				}
				if settings.ProvisionSettings.CreateDashboardURL && deployment.DashboardURL() == nil {
					t.Errorf("Expected deployment to have a value for DashboardURL but got nil")
					return
				}
				if !settings.ProvisionSettings.CreateDashboardURL && deployment.DashboardURL() != nil {
					t.Errorf("Expected deployment to have nil for DashboardURL but got an actual value")
					return
				}
				if settings.ProvisionSettings.CreateMetadata && deployment.Metadata() == nil {
					t.Errorf("Expected deployment to have a value for Metadata but got nil")
				}
				if !settings.ProvisionSettings.CreateMetadata && deployment.Metadata() != nil {
					t.Errorf("Expected deployment to have nil for Metadata but got an actual value")
				}
				if instancesRetrievable && deployment.FetchResponse() == nil {
					t.Errorf("Expected deployment to have a value for fetchResponse but got nil")
				}
				if !instancesRetrievable && deployment.FetchResponse() != nil {
					t.Errorf("Expected deployment to have nil for fetchResponse but got an actual value")
				}
				instancesRetrievable = !instancesRetrievable
			}
			settings.ProvisionSettings.CreateDashboardURL = !settings.ProvisionSettings.CreateDashboardURL
		}
		settings.ProvisionSettings.CreateMetadata = !settings.ProvisionSettings.CreateMetadata
	}
}

func TestDoOperation(t *testing.T) {
	//values := [3]bool{nil, false, true}

}
