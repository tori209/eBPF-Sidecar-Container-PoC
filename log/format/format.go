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
	RunnerStart
	RunnerFinish
	MessageReceived
	MessageBroken
	InternalError
)

// GOB를 통해 간단한 Manager->Executor 통신 구현을 위한 메세지 포맷
type TaskRequestMessage struct {
	JobID			uuid.UUID
	TaskID			uuid.UUID
	DataSourceURL	string
	DestinationURL	string
	RangeBegin		uint64			
	RangeEnd		uint64
	RunAsEvil		bool		// 실행 스크립트를 실제로 전송하기엔... Dummy 악성행위 생성용
}

type TaskResponseMessage struct {
	JobID			uuid.UUID
	TaskID			uuid.UUID
	Status			ReportType
}

// Executor 내 Runner->Watcher 상태 전파를 위한 메세지 포맷
type ReportMessage struct {
	Kind	ReportType
	JobID	uuid.UUID
	TaskID	uuid.UUID
}

// Watcher->Collector Metric 전달용 포맷
type L4Message struct {
	JobID		uuid.UUID
	TaskID		uuid.UUID
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
		msg.JobID.String(),
		msg.TaskID.String(),
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
