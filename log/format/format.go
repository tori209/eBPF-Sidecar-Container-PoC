package format

type ReportType int
const (
	JobStart	ReportType	= iota
	JobFinish
	JobFailed
	ServiceStart	
	ServiceFinish	
)

type ReportMessage struct {
	Kind	ReportType
	Data	[16]byte
}

type L4Message struct {
	MagicNumber	uint32 // 0xdeadb33f
	TaskID		string // UUID
	L4Metric
}

type L4Metric struct {
	TS		uint64
    IfIndex uint32
    SrcIP   uint32
    DstIP   uint32
	SPort	uint16
	DPort	uint16
	L4Proto	uint8
}
