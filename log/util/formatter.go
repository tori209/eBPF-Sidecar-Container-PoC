package util

import (
	"time"

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

func L4MessageToLogDao(msg *format.L4Message) (dao.LogDao) {
	return dao.LogDao{
		JobID: msg.JobID,
		TaskID: msg.TaskID,
		Timestamp: time.Unix(0, msg.TS), // TAI Time, due to ebpf, "bpf_ktime_get_tai_ns()"
		SrcIP: format.IpToString(msg.SrcIP),
		DstIP: format.IpToString(msg.DstIP),
		SrcPort: int32(msg.SPort),
		DstPort: int32(msg.DPort),
		Size: int64(msg.Size),
		L4Proto: format.GetL4ProtoString(msg.L4Proto),
	}
}

func MultipleL4MessageToLogDao(mm *[]format.L4Message) ([]*dao.LogDao) {
	ret := make([]*dao.LogDao, len(*mm))
	for idx, msg := range *mm {
		ld := L4MessageToLogDao(&msg)
		ret[idx] = &ld
	}
	return ret
}
