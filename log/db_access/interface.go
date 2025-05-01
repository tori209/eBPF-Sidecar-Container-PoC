package db_access

import (
	"container/list"
	"github.com/tori209/data-executor/log/format"
)

type TaskQueryRunner interface {
	SaveJob(*format.TaskRequestMessage)		error
	SaveJobs(*[]format.TaskRequestMessage)	error
	SaveTask(*format.TaskRequestMessage)	error
	SaveTasks(*[]format.TaskRequestMessage)	error
	SaveTasksFromList(*list.List)			error
}

type LogQueryRunner interface {

}
