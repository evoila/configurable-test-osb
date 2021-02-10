package controller

import (
	"fmt"
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
	var provisionRequest model.ProvisionRequest
	if err := context.ShouldBindJSON(&provisionRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while binding request body to struct",
			"error":   err.Error(),
		})
		return
	}
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\" or \"false\"",
		})
		return
	}

	fmt.Println(acceptsIncomplete)
	instanceID := context.Param("instance_id")
	fmt.Println(instanceID)
	//statuscode must be returned by ProvideService too
	statusCode, err := deploymentController.deploymentService.ProvideService(&provisionRequest, &instanceID,
		acceptsIncomplete == "true")
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, "in construction")
}
