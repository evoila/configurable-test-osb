package controller

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"net/http"
)

type BindingController struct {
	settings       *model.Settings
	bindingService *service.BindingService
}

func NewBindingController(bindingService *service.BindingService, settings *model.Settings) *BindingController {
	return &BindingController{
		settings:       settings,
		bindingService: bindingService,
	}
}

func (bindingController *BindingController) CreateBinding(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
	if bindingID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing url parameter \"binding_id\"",
			"error":   "invalid value, value must not be \"\" and unique",
		})
		return
	}
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
		})
		return
	}
	var bindingRequest model.CreateBindingRequest
	if err := context.ShouldBindBodyWith(&bindingRequest, binding.JSON); err != nil {
		log.Println("err was != nil")
		log.Println(err.Error())
		/*err = context.ShouldBindJSON(&bindingRequest)
		body := context.Request.Body
		b
		log.Println(err.Error())

		*/
		bindingController.rotateBinding(context)
		/*context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while binding request body to struct",
			"error":   err.Error(),
		})

		*/
		return
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(bindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}
	if bindingRequest.AppGUID != nil && *bindingRequest.AppGUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid app_guid in request, value must not be \"\"",
			"error":   "InvalidValue",
		})
		return
	}
	statusCode, response, err := bindingController.bindingService.CreateBinding(&bindingRequest, &instanceID, &bindingID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

func (bindingController *BindingController) rotateBinding(context *gin.Context) {
	log.Println("rotateBinding called")
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
	if bindingID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing url parameter \"binding_id\"",
			"error":   "invalid value, value must not be \"\" and unique",
		})
		return
	}
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
		})
		return
	}
	var rotateBindingRequest model.RotateBindingRequest
	log.Println(context)
	if err := context.ShouldBindBodyWith(&rotateBindingRequest, binding.JSON); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while binding request body to struct, the body must either be the one specified for creating bindings or the one for rotating bindings",
			"error":   err.Error(),
		})
		return
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(rotateBindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}
	statusCode, response, err := bindingController.bindingService.RotateBinding(&rotateBindingRequest, &instanceID, &bindingID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

func (bindingController *BindingController) FetchBinding(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
	serviceID := context.Query("service_id")
	planID := context.Query("plan_id")
	statusCode, response, err := bindingController.bindingService.FetchBinding(&instanceID, &bindingID, &serviceID, &planID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

func (bindingController *BindingController) CurrentBindings(context *gin.Context) {
	resp := struct {
		Bindings *map[string]*model.ServiceBinding `json:"service_bindings"`
	}{}
	resp.Bindings = bindingController.bindingService.CurrentBindings()
	//this won't show the fields inside a service instance, since it fields are not public (and not annotated)
	//only the names (with empty values) will be shown
	context.JSON(200, resp)
}
