package model

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"math/rand"
	"time"
)

type Catalog struct {
	//REQUIRED
	ServiceOfferings []ServiceOffering `json:"services"` //check if correct
}

func NewCatalog(catalogSettings *CatalogSettings) (*Catalog, error) {
	var catalog Catalog
	//var err error
	//create tags
	rand.Seed(time.Now().UnixNano())
	var tags []string
	for i := 0; i < catalogSettings.TagsMax; i++ {
		tag := generator.RandomString(4)
		for generator.ContainsString(tags, tag) {
			tag = generator.RandomString(4)
		}
		tags = append(tags, tag)
		//append(tags, RandomString(4))
	}
	for i := 0; i < catalogSettings.Amount; i++ {
		catalog.ServiceOfferings = append(catalog.ServiceOfferings, *newServiceOffering(catalogSettings, &catalog, tags))
	}
	return &catalog, nil
}

func (catalog *Catalog) GetServiceOfferingById(id string) *ServiceOffering {
	for _, offering := range catalog.ServiceOfferings {
		if id == offering.Id {
			return &offering
		}
	}
	return nil
}

func (catalog *Catalog) GetServiceOfferingByName(name string) *ServiceOffering {
	for _, offering := range catalog.ServiceOfferings {
		if name == offering.Name {
			return &offering
		}
	}
	return nil
}

func (catalog *Catalog) nameExists(name string) bool {
	for _, offering := range catalog.ServiceOfferings {
		if name == offering.Name {
			return true
		}
		for _, plan := range offering.Plans {
			if name == plan.Name {
				return true
			}
		}
	}
	return false
}
