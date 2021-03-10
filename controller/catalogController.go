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
	context.JSON(http.StatusOK, catalogController.catalogService.GetCatalog())
}

func (catalogController *CatalogController) GenerateCatalog(context *gin.Context) {
	settings, err := model.NewCatalogSettings()
	if err != nil {
		context.JSON(500, err)
		return
	}
	catalog, _ := model.NewCatalog(settings)

	context.JSON(http.StatusOK, catalog)
}
