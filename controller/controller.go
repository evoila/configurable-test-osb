package controller

import (
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

//FUNCTIONS FOR HANDLING ENDPOINT REQUESTS GO HERE

func Hello(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{"message": "Hello"})
}

func GetCatalog(context *gin.Context) {

}
//GET /v2/service_instances/:instance_id/last_operation
func LastOpServiceInstance(context *gin.Context)  {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}

	//VORERST
	context.String(http.StatusOK, uriParams.InstanceId)
}
//GET /v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation
func LastOpServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}

	//Vorerst

	fmt.Println(uriParams.InstanceId)
	fmt.Println("Help")
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//PUT /v2/service_instances/:instance_id
func ProvideService(context *gin.Context)  {

	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}

	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}

//GET /v2/service_instances/:instance_id
func FetchServiceInstance(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}

//PATCH /v2/service_instances/:instance_id
func UpdateServiceInstance(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}

//PUT /v2/service_instances/:instance_id/service_bindings/:binding_id
func CreateServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//PUT /v2/service_instances/:instance_id/service_bindings/:binding_id
func RotateServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//GET /v2/service_instances/:instance_id/service_bindings/:binding_id
func FetchServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id
func Unbind(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//DELETE /v2/service_instances/:instance_id
func Deprovide(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}