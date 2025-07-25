package dao

import (
	"github.com/google/uuid"
	"github.com/tori209/data-executor/log/format"
	"github.com/uptrace/bun"
)

type TaskDao struct {
	bun.BaseModel `bun:"table:tasks,alias:t"`

	JobID	uuid.UUID `bun:",pk,type:uuid"`
	TaskID	uuid.UUID `bun:",pk,type:uuid"`
	Source	format.DataSourceInfo `bun:"embed:src_"`
	Destination	format.DataSourceInfo `bun:"embed:dst_"`
	RangeBegin		int64
	RangeEnd		int64
	RunAsEvil		bool
}

