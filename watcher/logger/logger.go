package logger

import (
	"sync"
	"time"
	"net"
	"log"
	"encoding/gob"

	"github.com/tori209/data-executor/log/format"
)

type MetricManager struct {
	buffer			[]format.L4Message
	mu				*sync.Mutex
	timeTicker		*time.Ticker
	maxLen			uint32
	endpoint		string
	endpointType	string
}

func NewMetricManager (endpoint, endpointType string, size uint32) (*MetricManager) {
	mm := MetricManager {
		buffer: nil,
		mu:	new(sync.Mutex),
		timeTicker: nil,
		maxLen: size,
		endpoint: endpoint,
		endpointType: endpointType,
	}
	return &mm
}

func (mm *MetricManager) IsRunning() (bool) {
	// Check Mutex is Locked Before calling it.
	return mm.timeTicker != nil
}

func (mm *MetricManager) Run(d time.Duration) {
	mm.mu.Lock()

	if mm.IsRunning() {  
		mm.mu.Unlock()
		return
	}
	mm.timeTicker = time.NewTicker(d)

	mm.mu.Unlock()

	go func() {
		for _ = range mm.timeTicker.C {
			mm.flush()
		}
	}()
}

func (mm *MetricManager) Stop() {
	mm.mu.Lock()

	if !mm.IsRunning() {  return  }
	mm.timeTicker.Stop()
	mm.timeTicker = nil

	mm.mu.Unlock()

	mm.flush()
}

func (mm *MetricManager) Append(message format.L4Message) {
	mm.mu.Lock()
	mm.buffer = append(mm.buffer, message)

	if uint32(len(mm.buffer)) >= mm.maxLen {
		mm.mu.Unlock()
		go mm.flush()
	} else {
		mm.mu.Unlock()
	}
}

func (mm *MetricManager) flush() {
	mm.mu.Lock()
	if len(mm.buffer) == 0 {
		mm.mu.Unlock()
		return
	}

	logs := make([]format.L4Message, len(mm.buffer))
	copy(logs, mm.buffer)
	mm.buffer = mm.buffer[:0]
	mm.mu.Unlock()

	go mm.sendToServer(logs)
}

func (mm *MetricManager) sendToServer(logs []format.L4Message) {
	conn, err := net.Dial(mm.endpointType, mm.endpoint)
	if err != nil {
		log.Printf("Watcher/Logger: Failed to Connect Collector: %+v", err)
		return
	}
	defer conn.Close()

	// UUID를 추가해줘야...
	enc := gob.NewEncoder(conn)
	if err := enc.Encode(logs); err != nil {
		log.Printf("Watcher/Logger: Failed to Send Log: %+v", err)
	}
}
