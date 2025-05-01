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
	MessageFailed
)

func (rt *ReportType) String() string {
	switch *rt {
	case	TaskStart:
		return "TaskStart"
	case	TaskFinish:
		return "TaskFinish"
	case	TaskFailed:
		return "TaskFailed"
	case	JobStart:
		return "JobStart"
	case	JobFinish:
		return "JobFinish"
	case	JobFailed:
		return "JobFailed"
	case	RunnerStart:
		return "RunnerStart"
	case	RunnerFinish:
		return "RunnerFinish"
	case	MessageReceived:
		return "MessageReceived"
	case	MessageFailed:
		return "MessageFailed"
	default:
		return "UNKNOWN"
	}
}

type DataSourceInfo struct {
	Endpoint	string
	BucketName	string
	ObjectName	string
}

// GOB를 통해 간단한 Manager->Executor 통신 구현을 위한 메세지 포맷
type TaskRequestMessage struct {
	JobID			uuid.UUID
	TaskID			uuid.UUID
	DataSource		DataSourceInfo
	DestinationURL	string
	RangeBegin		int64			
	RangeEnd		int64
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
	TS		int64
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

func (msg *L4Message) GetSrcAsString() string {
	return fmt.Sprintf("%s:%d", IpToString(msg.SrcIP), msg.SPort)
}

func (msg *L4Message) GetDstAsString() string {
	return fmt.Sprintf("%s:%d", IpToString(msg.DstIP), msg.DPort)
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
