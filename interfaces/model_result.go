package interfaces

import "go.mongodb.org/mongo-driver/bson/primitive"

type Result interface {
	Value() map[string]any
	SetValue(key string, value any)
	GetValue(key string) (value any)
	GetTaskId() (id primitive.ObjectID)
	SetTaskId(id primitive.ObjectID)
}
