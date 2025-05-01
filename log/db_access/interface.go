package db_access

import (
	"github.com/tori209/data-executor/log/format"
)

type task_db interface {
	SaveJobs(*[]format.TaskRequestMessage) error
	SaveTasks(*[]format.TaskRequestMessage) error
}
