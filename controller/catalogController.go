package controller

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CatalogController struct {
	catalogService *service.CatalogService
	settings       *model.Settings
}

func NewCatalogController(catalogService *service.CatalogService, settings *model.Settings) CatalogController {
	return CatalogController{
		catalogService: catalogService,
		settings:       settings,
	}
}

func (catalogController *CatalogController) GetCatalog(context *gin.Context) {
	_, err := bindAndCheckHeader(context, catalogController.settings)
	if err != nil {
		//context.json here or in bindAndCheck?
	} else {
		context.JSON(http.StatusOK, catalogController.catalogService.GetCatalog())
	}
}
