// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql

import (
	model "github.com/joincivil/civil-events-processor/pkg/model"
)

type ArticlePayload struct {
	Key   string                    `json:"key"`
	Value model.ArticlePayloadValue `json:"value"`
}
type BlockData struct {
	BlockNumber int    `json:"blockNumber"`
	TxHash      string `json:"txHash"`
	TxIndex     int    `json:"txIndex"`
	BlockHash   string `json:"blockHash"`
	Index       int    `json:"index"`
}
type DateRange struct {
	Gt *int `json:"gt"`
	Lt *int `json:"lt"`
}
type KycCreateApplicantInput struct {
	FirstName          string  `json:"firstName"`
	LastName           string  `json:"lastName"`
	Email              *string `json:"email"`
	MiddleName         *string `json:"middleName"`
	Profession         *string `json:"profession"`
	Nationality        *string `json:"nationality"`
	CountryOfResidence *string `json:"countryOfResidence"`
	DateOfBirth        *string `json:"dateOfBirth"`
	BuildingNumber     *string `json:"buildingNumber"`
	Street             *string `json:"street"`
	AptNumber          *string `json:"aptNumber"`
	City               *string `json:"city"`
	State              *string `json:"state"`
	Zipcode            *string `json:"zipcode"`
}
type Metadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
