package interfaces

import (
	"github.com/crawlab-team/crawlab-db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ModelBaseService interface {
	GetModelId() (id ModelId)
	SetModelId(id ModelId)
	GetById(id primitive.ObjectID) (res Model, err error)
	Get(query bson.M, opts *mongo.FindOptions) (res Model, err error)
	GetList(query bson.M, opts *mongo.FindOptions) (res List, err error)
	DeleteById(id primitive.ObjectID, args ...any) (err error)
	Delete(query bson.M, args ...any) (err error)
	DeleteList(query bson.M, args ...any) (err error)
	ForceDeleteList(query bson.M, args ...any) (err error)
	UpdateById(id primitive.ObjectID, update bson.M, args ...any) (err error)
	Update(query bson.M, update bson.M, fields []string, args ...any) (err error)
	UpdateDoc(query bson.M, doc Model, fields []string, args ...any) (err error)
	Insert(u User, docs ...any) (err error)
	Count(query bson.M) (total int, err error)
}

type ModelService interface {
	GetBaseService(id ModelId) (svc ModelBaseService)
}
