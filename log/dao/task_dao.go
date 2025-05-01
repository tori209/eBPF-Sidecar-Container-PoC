package dao

import (
	"github.com/tori209/data-executor/log/format"
	"github.com/uptrace/bun"
	"github.com/google/uuid"
)

type TaskDao struct {
	bun.BaseModel `bun:"table:tasks,alias:t"`

	JobID	uuid.UUID `bun:",pk,type:uuid"`
	TaskID	uuid.UUID `bun:",pk,type:uuid"`
	Source	format.DataSourceInfo `bun:"embed:src_"`
	DestinationURL	string
	RangeBegin		int64
	RangeEnd		int64
}

