package dao

import (
	"github.com/tori209/data-executor/log/format"
	"github.com/uptrace/bun"
	"github.com/google/uuid"
)

type JobDao struct {
	bun.BaseModel `bun:"table:jobs,alias:j"`

	JobID	uuid.UUID `bun:",pk,type:uuid"`
	Source	format.DataSourceInfo `bun:"embed:src_"`
	DestinationURL	string
	RangeBegin		int64
	RangeEnd		int64
}

