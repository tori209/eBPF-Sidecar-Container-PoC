package bpf

import (
  "os"
  "os/signal"
  "sync"
  "net"
  "log"
  "time"
  "errors"
  "encoding/binary"
  "bytes"

  "github.com/tori209/data-executor/log/format"
  "github.com/tori209/data-executor/watcher/logger"

  "github.com/cilium/ebpf"
  "github.com/cilium/ebpf/link"
  "github.com/cilium/ebpf/ringbuf"
)

const DEFAULT_IFACE_NAME="eth0"

type BpfTrafficCapture struct {
	mu			*sync.Mutex	
	logManager	*logger.MetricManager
	jobID		[16]byte
	taskID		[16]byte
}

func NewBpfTrafficCapture(endpoint, endpointType string, size uint32) (* BpfTrafficCapture) {
	return &BpfTrafficCapture{
		mu:	new(sync.Mutex),
		logManager: logger.NewMetricManager(endpoint, endpointType, size),
	}
}

func (btc *BpfTrafficCapture) SetJobID(newId [16]byte) {
	btc.mu.Lock()
	defer btc.mu.Unlock()
	btc.jobID = newId
}

func (btc *BpfTrafficCapture) SetTaskID(newId [16]byte) {
	btc.mu.Lock()
	defer btc.mu.Unlock()
	btc.taskID = newId
}

func (btc *BpfTrafficCapture) StartCapture(ifaceName string) {
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
	prog := coll.Programs["egress_capture"]
    if prog == nil {
        log.Fatalf("program not found")
    }
	lnk, err := link.AttachTCX(link.TCXOptions{
		Interface:	iface.Index,
		Program: 	prog,
		Attach: 	ebpf.AttachTCXEgress,
	})
		/*
		XDP(link.XDPOptions{
			Program:	prog,
			Interface:	iface.Index,
			Flags:		link.XDPGenericMode,
		})
		*/
	if err != nil {
		log.Fatalf("Failed to attach TCXEgress to %s: %v",
			ifaceName,
			err,
		)
	}
	defer lnk.Close()

	// Get RingBuf Reference =======================================
	rb, err := ringbuf.NewReader(coll.Maps["egress_events"])
    if err != nil {
        log.Fatalf("failed to open ring buffer: %v", err)
    }
	rb.SetDeadline(time.Unix(1,0)) // 읽기 Deadline = 1 sec
    defer rb.Close()

	// Start Packet Capture
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	btc.logManager.Run(1 * time.Second)
	go btc.receive_from_ringbuf(rb)

	<- sig
	
	btc.logManager.Stop()
}

func (btc *BpfTrafficCapture) receive_from_ringbuf(rb *ringbuf.Reader) {
	for {
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
			JobID: btc.jobID,
			TaskID: btc.taskID,
		}
		btc.mu.Unlock()
		btc.logManager.Append(msg)
	}
}
