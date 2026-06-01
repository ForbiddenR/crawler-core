package interfaces

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PluginService interface {
	Module
	SetFsPathBase(path string)
	SetMonitorInterval(interval time.Duration)
	InstallPlugin(id primitive.ObjectID) (err error)
	UninstallPlugin(id primitive.ObjectID) (err error)
	StartPlugin(id primitive.ObjectID) (err error)
	StopPlugin(id primitive.ObjectID) (err error)
	GetPublicPluginList() (res any, err error)
	GetPublicPluginInfo(fullName string) (res any, err error)
}
