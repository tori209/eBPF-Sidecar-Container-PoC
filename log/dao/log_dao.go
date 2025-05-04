package dao

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/google/uuid"
)

type LogDao struct {
	bun.BaseModel `bun:"table:logs,alias:l"`
	
	LogID	int64			`bun:"id,pk,autoincrement"`
	JobID	uuid.UUID		`bun:"type:uuid"`
	TaskID	uuid.UUID		`bun:"type:uuid"`
	Timestamp	time.Time	`bun:"timestamp,notnull"`
	SrcIP	string			`bun:"src_ip,type:inet"`
	DstIP	string			`bun:"dst_ip,type:inet"`
	SrcPort	int32			`bun:"src_port"` // PostgreSQL이 Unsigned 미지원하여 Signed로 확장.
	DstPort	int32			`bun:"dst_port"`
	Size	int64			`bun:"packet_size"`
	L4Proto	string			`bun:"protocol"`
}
