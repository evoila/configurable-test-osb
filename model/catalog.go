package model

import (
	"github.com/evoila/configurable-test-osb/generator"
	"math/rand"
	"time"
)

type Catalog struct {
	//REQUIRED
	ServiceOfferings *[]*ServiceOffering `json:"services" validate:"required,dive"` //check if correct
}

//NewCatalog generates a new randomized catalog regarding to the catalogSettings.
//Returns *Catalog (the newly generated catalog)
func NewCatalog(catalogSettings *CatalogSettings) *Catalog {
	var offerings []*ServiceOffering
	var catalog = Catalog{ServiceOfferings: &offerings}
	rand.Seed(time.Now().UnixNano())
	var tags []string
	for i := 0; i < catalogSettings.TagsMax; i++ {
		tag := generator.RandomString(4)
		for generator.ContainsString(tags, tag) {
			tag = generator.RandomString(4)
		}
		tags = append(tags, tag)
	}
	for i := 0; i < catalogSettings.Amount; i++ {
		*catalog.ServiceOfferings = append(*catalog.ServiceOfferings, newServiceOffering(catalogSettings, &catalog, tags))
	}
	return &catalog
}

func (catalog *Catalog) GetServiceOfferingById(id string) (*ServiceOffering, bool) {
	for _, offering := range *catalog.ServiceOfferings {
		if id == offering.ID {
			return offering, true
		}
	}
	return nil, false
}

func (catalog *Catalog) GetServiceOfferingByName(name string) *ServiceOffering {
	for _, offering := range *catalog.ServiceOfferings {
		if name == offering.Name {
			return offering
		}
	}
	return nil
}

func (catalog *Catalog) nameExists(name string) bool {
	for _, offering := range *catalog.ServiceOfferings {
		if name == offering.Name {
			return true
		}
		for _, plan := range *offering.Plans {
			if name == plan.Name {
				return true
			}
		}
	}
	return false
}
