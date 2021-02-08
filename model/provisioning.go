package model

type provisionRequest struct {
	ServiceID        string          `json:"service_id" binding:"required"`
	PlanID           string          `json:"plan_id" binding:"required"`
	Context          interface{}     `json:"context"`
	OrganizationGUID string          `json:"organization_guid" binding:"required"`
	SpaceGUID        string          `json:"space_guid" binding:"required"`
	Parameters       interface{}     `json:"parameters"`
	MaintenanceInfo  MaintenanceInfo `json:"maintenance_info"`
}

type provisionResponse struct {
	DashboardURL string                  `json:"dashboard_url,omitempty"`
	Operation    string                  `json:"operation,omitempty"`
	Metadata     ServiceInstanceMetadata `json:"metadata,omitempty"`
}
