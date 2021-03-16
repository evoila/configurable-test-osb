package service

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"log"
)

type CatalogService struct {
	catalog *model.Catalog
}

func NewCatalogService(catalog *model.Catalog) CatalogService {
	return CatalogService{
		catalog: catalog,
	}
}

//GetCatalog() returns the catalog used by this service broker.
//Returns *model.Catalog
func (catalogService *CatalogService) GetCatalog() *model.Catalog {
	return catalogService.catalog
}

func (catalogService *CatalogService) logCatalog() {
	s, _ := json.MarshalIndent(catalogService.GetCatalog(), "", "\t")
	log.Print(string(s))
}
