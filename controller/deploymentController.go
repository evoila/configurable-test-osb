package controller

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DeploymentController struct {
	settings          *model.Settings
	deploymentService *service.DeploymentService
}

func NewDeploymentController(deploymentService *service.DeploymentService, settings *model.Settings) *DeploymentController {
	return &DeploymentController{
		settings:          settings,
		deploymentService: deploymentService,
	}
}

func (deploymentController *DeploymentController) Provision(context *gin.Context) {
	var provisionRequest model.ProvideServiceInstanceRequest
	if err := context.ShouldBindJSON(&provisionRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while binding request body to struct",
			"error":   err.Error(),
		})
		return
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(provisionRequest.Parameters)
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
			//fmt.Println(acceptsIncomplete)
			//fmt.Println(instanceID)
			//statuscode must be returned by ProvideService too
		})
		return
	}

	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}

	instanceID := context.Param("instance_id")
	if provisionRequest.OrganizationGUID == "" {
		context.JSON(http.StatusBadRequest, &model.ServiceBrokerError{
			Error:       "EmptyOrganizationGUID",
			Description: "organization_guid must be a non-empty string",
		})
		return
	}
	if provisionRequest.SpaceGUID == "" {
		context.JSON(http.StatusBadRequest, &model.ServiceBrokerError{
			Error:       "EmptySpaceGUID",
			Description: "space_guid must be a non-empty string",
		})
		return
	}

	statusCode, response, err := deploymentController.deploymentService.ProvideService(&provisionRequest, &instanceID) //accepts_incomplete used to be here but is not needed
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

func (deploymentController *DeploymentController) FetchServiceInstance(context *gin.Context) {
	instanceID := context.Param("instance_id")
	serviceID := context.Query("service_id")
	planID := context.Query("plan_id")
	statusCode, response, err := deploymentController.deploymentService.FetchServiceInstance(&instanceID, &serviceID, &planID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

func (deploymentController *DeploymentController) UpdateServiceInstance(context *gin.Context) {
	//what happens if empty???
	instanceID := context.Param("instance_id")
	var updateRequest model.UpdateServiceInstanceRequest
	if err := context.ShouldBindJSON(&updateRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while binding request body to struct",
			"error":   err.Error(),
		})
		return
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(updateRequest.Parameters)
	//GROUP IN ITS OWN FUNCTION???
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
		})
		return
	}
	//checking for nil not needed because asyncEndpoint gets default value = false
	//if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
	if *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}

	statusCode, response, err := deploymentController.deploymentService.UpdateServiceInstance(&updateRequest, &instanceID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}
