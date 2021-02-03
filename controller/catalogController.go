package controller

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CatalogController struct {
	catalogService *service.CatalogService
}

func NewCatalogController(catalogService *service.CatalogService) CatalogController {
	return CatalogController{catalogService: catalogService}
}

func (catalogController *CatalogController) GetCatalog(context *gin.Context) {
	err := bindAndCheckHeader(context)
	if err != nil {
		//context.json here or in bindAndCheck?
	} else {
		context.JSON(http.StatusOK, catalogController.catalogService.GetCatalog())
	}
}
