package entity

type FilterSelectOption struct {
	Value any    `json:"value" bson:"value"`
	Label string `json:"label" bson:"label"`
}
