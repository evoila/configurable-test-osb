package model

type Catalog struct {
	ServiceOfferings []ServiceOffering `json:"services" binding:"required"` //check if correct
}
