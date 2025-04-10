package bpf

import (
  "os"
  "sync"
  "net"
  "log"
  "time"
  "errors"
  "encoding/binary"
  "bytes"

  "github.com/tori209/data-executor/log/format"
  "github.com/tori209/data-executor/watcher/logger"

  "github.com/google/uuid"

  "github.com/cilium/ebpf"
  "github.com/cilium/ebpf/link"
  "github.com/cilium/ebpf/ringbuf"
)

type BpfTrafficCapture struct {
	isRunning	bool
	mu			*sync.Mutex	
	logManager	*logger.MetricManager
	stopChan	chan bool
	JobID		uuid.UUID
	TaskID		uuid.UUID
}

func NewBpfTrafficCapture(endpoint, endpointType string, size uint32) (* BpfTrafficCapture) {
	log.Printf("Create New TrafficCapture\n");
	return &BpfTrafficCapture{
		isRunning: false,
		mu:	new(sync.Mutex),
		logManager: logger.NewMetricManager(endpoint, endpointType, size),
		stopChan: make(chan bool),
	}
}

// RPC 요청 전용으로 사용할 예정
func (btc *BpfTrafficCapture) ReceiveTaskMessage(rm *format.ReportMessage, res *format.ReportMessage) error {
	res.Kind = format.MessageReceived
	res.JobID = rm.JobID
	res.TaskID = rm.TaskID

	btc.mu.Lock()
	defer btc.mu.Unlock()

	switch rm.Kind {
	case format.TaskStart:
		btc.JobID = rm.JobID
		btc.TaskID = rm.TaskID
		log.Printf("[BpfTrafficCapture] Task(JobID: %s, TaskID: %s) Started.",
			btc.JobID.String(),
			btc.TaskID.String(),
		)

		return nil
	case format.TaskFinish, format.TaskFailed:
		log.Printf("[BpfTrafficCapture] Task(JobID: %s, TaskID: %s) Finished.",
			rm.JobID.String(),
			rm.TaskID.String(),
		)
		btc.JobID = uuid.Nil
		btc.TaskID = uuid.Nil
		return nil
	case format.RunnerStart:
		log.Printf("[BpfTrafficCapture] Runner Started.")
		return nil
	case format.RunnerFinish:
		log.Printf("[BpfTrafficCapture] Runner Stopped.")
		return nil
	default:
		log.Printf("[BpfTrafficCapture] Unexpected Message Received: %s", rm.Kind.String())
		res.Kind = format.MessageFailed
		return nil
	}
}

func (btc *BpfTrafficCapture) StartCapture(ifaceName string) {
	btc.mu.Lock()
	if btc.isRunning {
		btc.mu.Unlock()
		return
	}
	btc.isRunning = true
	btc.mu.Unlock()

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("Failed to find Interface %s: %+v", ifaceName, err)
	}
	spec, err := ebpf.LoadCollectionSpec("tc_capture.bpf.o")
	if err != nil {
		log.Fatalf("failed to load spec: %+v", err)
	}

	coll, err := ebpf.NewCollection(spec)
    if err != nil {
        log.Fatalf("failed to create collection: %v", err)
    }
    defer coll.Close()
	
	// Attach TCXEgress =============================================
	prog := coll.Programs["tc_egress_capture"]
    if prog == nil {
        log.Fatalf("program '%s' not exists", "tc_egress_capture") // 나중에 바꾸던가.
    }
	lnk, err := link.AttachTCX(link.TCXOptions{
		Interface:	iface.Index,
		Program: 	prog,
		Attach: 	ebpf.AttachTCXEgress,
	})
	if err != nil {
		log.Fatalf("Failed to attach TCXEgress to %s: %v",
			ifaceName,
			err,
		)
	}
	defer lnk.Close()

	btc.logManager.Run(1 * time.Second)
	
	btc.receive_from_ringbuf(coll)
}

func (btc *BpfTrafficCapture) StopCapture() {
	btc.mu.Lock()
	if !btc.isRunning {
		btc.mu.Unlock()
		return
	}
	btc.isRunning = false
	btc.stopChan <- true
	btc.logManager.Stop()
	btc.mu.Unlock()
}

func (btc *BpfTrafficCapture) receive_from_ringbuf(coll *ebpf.Collection) {
	// Get RingBuf Reference =======================================
	rb, err := ringbuf.NewReader(coll.Maps["egress_metrics"])
    if err != nil {
        log.Fatalf("failed to open ring buffer: %v", err)
    }
	rb.SetDeadline(time.Unix(1,0)) // 읽기 Deadline = 1 sec
    defer rb.Close()

	for {
		select {
		case <-btc.stopChan:
			break	
		default:
		record, err := rb.Read() 
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		} else if err != nil {
			log.Fatalf("failed to read ringbuf: %v", err)
        }

		var e format.L4Metric
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &e); err != nil {
			log.Printf("failed to parse event: %v", err)
			continue
		}

		btc.mu.Lock()
		msg := format.L4Message{
			L4Metric: e,
			JobID: btc.JobID,
			TaskID: btc.TaskID,
		}
		btc.mu.Unlock()
		btc.logManager.Append(&msg)
		}
	}
}
