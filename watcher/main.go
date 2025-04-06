package main

import (
  "log"
  "os"
  "net"
  "net/rpc"

  //"github.com/tori209/data-executor/log/format"
  "github.com/tori209/data-executor/watcher/bpf"
)

func main() {
	
	if os.Getenv("EXECUTOR_ENV_READY") != "Ready" {
		log.Fatalf("Environment is not set. Apply ConfigMap in config/executor-pod.yaml.")
	}

	log.Printf("Start Initialization.")
	socketPath := os.Getenv("WATCHER_SOCK_PATH")
	if socketPath == "" {
		log.Fatalf("WATCHER_SOCK_PATH not found in env")
	}
	collectorSocketPath := os.Getenv("COLLECTOR_SOCK_PATH")
	if collectorSocketPath == "" {
		log.Fatalf("COLLECTOR_SOCK_PATH not found in env")
	}
	targetInterfaceName := os.Getenv("TARGET_INTERFACE_NAME")
	if targetInterfaceName == "" {
		log.Fatalf("TARGET_INTERFACE_NAME not found in env")
	}

	log.Printf("Create Watcher UDS")
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
	log.Printf("Watcher UDS created in %s.", socketPath)

	log.Printf("Create/Load BPF Capture")
	capture := bpf.NewBpfTrafficCapture(
		collectorSocketPath,
		"unix",
		uint32(1024),
	)
	go capture.StartCapture(targetInterfaceName)
	log.Printf("[Watcher] Capture Started.")

	// Runner와의 통신.
	rpc.Register(capture)
	for {
		conn, err := watcher_listener.Accept();
		if err != nil {
			log.Printf("Connection Error: %+v\n", err)
			continue
		}
		rpc.ServeConn(conn)
	}
	log.Printf("[Watcher] Capture Finished.")
	capture.StopCapture()
}

