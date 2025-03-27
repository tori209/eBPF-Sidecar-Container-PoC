package main

import (
  "io"
  "log"
  "os"
  "net"
  "encoding/gob"

  "github.com/google/uuid"
  "github.com/tori209/data-executor/log/format"
)

// collectorSocketPath = os.Getenv("COLLECTOR_SOCK_PATH")

func main() {
	if (os.Getenv("EXECUTOR_ENV_READY") != "Ready") {
		log.Fatalf("Environment is not set. Apply ConfigMap in config/executor-pod.yaml.")
	}

	socketPath := os.Getenv("WATCHER_SOCK_PATH")
	if (socketPath == "") {
		log.Fatalf("Env Not Set")
	}

	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}
	
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to create Socket: %+v", err)
	}
	defer listener.Close()
	log.Printf("Start")

	for {
		conn, err := listener.Accept();
		if err != nil {
			log.Printf("Connection Error: %+v\n", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
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
		u:=uuid.UUID(msg.Data)
		log.Printf("%s / Type: %d", u.String(), msg.Kind)
	}
}
