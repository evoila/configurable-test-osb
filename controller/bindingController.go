package controller

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"net/http"
	"time"
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
	if instanceID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing url parameter \"instance_id\"",
			"error":   "invalid value, value must not be \"\" and unique",
		})
		return
	}
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
		//checking, if the request was a binding request and if so, rotate the bindings
		bindingController.rotateBinding(context)
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

func (bindingController *BindingController) PollOperationState(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
	var serviceID *string
	valueServiceID, exists := context.GetQuery("service_id")
	if exists {
		log.Printf("service_id valueServiceID: %v\n", valueServiceID)
		if valueServiceID == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"service_id\" must not be an empty string (but can be omitted)",
			})
		} else {
			serviceID = &valueServiceID
			log.Printf("service_id assigned valueServiceID: %v\n", *serviceID)
		}
	}
	var planID *string
	valuePlanID, exists := context.GetQuery("plan_id")
	if exists {
		if valuePlanID == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"plan_id\" must not be an empty string (but can be omitted)",
			})
		} else {
			planID = &valuePlanID
		}
	}
	var operation *string
	valueOperation, exists := context.GetQuery("operation")
	if exists {
		if valueOperation == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"operation\" must not be an empty string (but can be omitted)",
			})
		} else {
			operation = &valueOperation
		}
	}
	statusCode, response, err := bindingController.bindingService.PollOperationState(&instanceID, &bindingID, serviceID, planID, operation)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	if response.State == model.PROGRESSING && bindingController.settings.BindingSettings.RetryPollBindingOperationAfterSeconds > 0 {
		retryAfter := time.Second * time.Duration(bindingController.settings.BindingSettings.RetryPollBindingOperationAfterSeconds)
		context.Header("Retry-After", retryAfter.String())
	}
	context.JSON(statusCode, response)
}

func (bindingController *BindingController) Unbind(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")

	serviceOfferingID, exists := context.GetQuery("service_id")
	if !exists {
		context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
			Error:       "MalformedRequest",
			Description: "service_id must be included as query parameter",
		})
		return
	}
	servicePlanID, exists := context.GetQuery("plan_id")
	if !exists {
		context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
			Error:       "MalformedRequest",
			Description: "plan_id must be included as query parameter",
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
	var deleteRequest model.DeleteRequest
	_ = context.ShouldBindBodyWith(&deleteRequest, binding.JSON)
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(deleteRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}
	statusCode, response, err := bindingController.bindingService.Unbind(&deleteRequest, &instanceID, &bindingID, &serviceOfferingID, &servicePlanID)
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
