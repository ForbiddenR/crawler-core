package interfaces

import (
	"time"

	"github.com/crawlab-team/crawlab-db/generic"
)

type ResultService interface {
	Insert(records ...any) (err error)
	List(query generic.ListQuery, opts *generic.ListOptions) (results []any, err error)
	Count(query generic.ListQuery) (n int, err error)
	Index(fields []string)
	SetTime(t time.Time)
	GetTime() (t time.Time)
}
