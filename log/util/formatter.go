package util

import (
	"github.com/tori209/data-executor/log/dao"
	"github.com/tori209/data-executor/log/format"
)

func TaskRequestToJobDao(tr *format.TaskRequestMessage) (dao.JobDao) {
	return dao.JobDao{
		JobID: tr.JobID,
		Source: tr.DataSource,
		DestinationURL: tr.DestinationURL,
		RangeBegin: tr.RangeBegin,
		RangeEnd: tr.RangeEnd,
	}
}

func TaskRequestToTaskDao(tr *format.TaskRequestMessage) (dao.TaskDao) {
	return dao.TaskDao{
		JobID: tr.JobID,
		TaskID: tr.TaskID,
		Source: tr.DataSource,
		DestinationURL: tr.DestinationURL,
		RangeBegin: tr.RangeBegin,
		RangeEnd: tr.RangeEnd,
	}
}
