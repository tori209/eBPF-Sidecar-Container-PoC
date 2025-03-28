package format

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
