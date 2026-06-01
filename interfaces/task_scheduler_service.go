package interfaces

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskSchedulerService interface {
	TaskBaseService
	// Enqueue task into the task queue
	Enqueue(t Task) (t2 Task, err error)
	// Cancel task to corresponding node
	Cancel(id primitive.ObjectID, args ...any) (err error)
	// SetInterval set the interval or duration between two adjacent fetches
	SetInterval(interval time.Duration)
}
