package format

import (
	"fmt"
	"encoding/binary"
	"net"
	"github.com/google/uuid"
)

type ReportType int
const (
	TaskStart	ReportType	= iota
	TaskFinish
	TaskFailed
	JobStart
	JobFinish
	JobFailed
	ServiceStart	
	ServiceFinish	
)

type ReportMessage struct {
	Kind	ReportType
	JobID	[16]byte
	TaskID	[16]byte
}

type L4Message struct {
	JobID		[16]byte // UUID
	TaskID		[16]byte // UUID
	L4Metric
}

type L4Metric struct {
	TS		uint64
	SrcIP   uint32
	DstIP   uint32
	SPort	uint16
	DPort	uint16
	Size	uint32
	L4Proto	uint8
}

func (msg *L4Message) String() string {
	return fmt.Sprintf(
		"%s\tJobID: %s, TaskID: %s",
		msg.L4Metric.String(),
		uuid.UUID(msg.JobID).String(),
		uuid.UUID(msg.TaskID).String(),
	)
}

func (msg *L4Metric) String() string {
	return fmt.Sprintf(
		"[%d] (%s) %s:%d -> %s:%d (Size: %d)",
		msg.TS,
		GetL4ProtoString(msg.L4Proto),
		IpToString(msg.SrcIP),
		msg.SPort,
		IpToString(msg.DstIP),
		msg.DPort,
		msg.Size,
	)
}

func GetL4ProtoString(proto_id uint8) string {
	switch proto_id {
	case 0:
		return "IP"
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	default:
		return "Unknown"
	}
}

func IpToString(ip uint32) string {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, ip)
	return net.IP(b).String()
}
