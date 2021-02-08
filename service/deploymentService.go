package service

import "github.com/MaxFuhrich/serviceBrokerDummy/model"

type DeploymentService struct {
	catalog *model.Catalog
	//pointer to settings?
	serviceInstances *map[string]model.ServiceDeployment
}

func NewDeploymentService(catalog *model.Catalog, serviceInstances *map[string]model.ServiceDeployment) *DeploymentService {
	return &DeploymentService{
		catalog:          catalog,
		serviceInstances: serviceInstances,
	}
}
