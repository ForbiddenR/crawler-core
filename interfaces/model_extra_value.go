package interfaces

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExtraValue interface {
	Model
	GetValue() (v any)
	SetValue(v any)
	GetObjectId() (oid primitive.ObjectID)
	SetObjectId(oid primitive.ObjectID)
	GetModel() (m string)
	SetModel(m string)
	GetType() (t string)
	SetType(t string)
}
