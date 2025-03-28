package main

import (
  "io"
  "log"
  "os"
  "net"
  "encoding/gob"

  "github.com/tori209/data-executor/log/format"
  "github.com/tori209/data-executor/watcher/bpf"
)

// collectorSocketPath = os.Getenv("COLLECTOR_SOCK_PATH")

func main() {
	if os.Getenv("EXECUTOR_ENV_READY") != "Ready" {
		log.Fatalf("Environment is not set. Apply ConfigMap in config/executor-pod.yaml.")
	}

	socketPath := os.Getenv("WATCHER_SOCK_PATH")
	if socketPath == "" {
		log.Fatalf("Env Not Set")
	}
	collectorSocketPath := os.Getenv("COLLECTOR_SOCK_PATH")
	if collectorSocketPath == "" {
		log.Fatalf("COLLECTOR_SOCK_PATH not found in env")
	}

	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}
	
	// Socket 생성
	watcher_listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to create Socket: %+v", err)
	}
	defer os.Remove(socketPath)
	defer watcher_listener.Close()
	log.Printf("Start")

	capture := bpf.NewBpfTrafficCapture(
		collectorSocketPath,
		"unix",
		uint32(1024),
	)

	// Runner와의 통신.
	for {
		conn, err := watcher_listener.Accept();
		if err != nil {
			log.Printf("Connection Error: %+v\n", err)
			continue
		}
		go handleConn(conn, capture)
	}
}

func handleConn(conn net.Conn, capture *bpf.BpfTrafficCapture) {
	defer conn.Close()

	dec := gob.NewDecoder(conn)
	var msg format.ReportMessage
	for {
		err := dec.Decode(&msg)
		if err == io.EOF {
			log.Printf("Connection Closed")
			break
		}
		if err != nil {
			log.Printf("Read Failed: %+v\n", err)
			break
		}

		// Case Handling
		if msg.Kind == format.TaskStart {
			capture.SetTaskID(msg.TaskID)
			continue
		}

		if msg.Kind == format.TaskFinish {
			capture.SetTaskID([16]byte{})
			continue
		}

		if msg.Kind == format.JobStart {
			capture.SetJobID(msg.JobID)
			continue
		}

		if msg.Kind == format.JobFinish {
			capture.SetJobID([16]byte{})
			continue
		}

		if msg.Kind == format.ServiceStart {
			log.Printf("Runner Service Started\n")
			continue
		}
		
		if msg.Kind == format.ServiceFinish {
			log.Printf("Runner Service Finished. Terminating...\n")
			break
		}
	}
}
