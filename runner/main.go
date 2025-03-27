package main

import (
    "encoding/gob"
	"log"
    "net"
    "time"
	"os"

	"github.com/google/uuid"
	"github.com/tori209/data-executor/log/format"
)

func main() {
	socketPath := os.Getenv("WATCHER_SOCK_PATH")

	var conn net.Conn
	for {
		if con, err := net.Dial("unix", socketPath); err == nil {
			conn = con
			break
		}
		log.Printf("Failed to connect socket %s. Retry...", socketPath)
		time.Sleep(1 * time.Second)
	}
	defer conn.Close()

	enc := gob.NewEncoder(conn)
	
	var msg format.ReportMessage
	msgUUID := uuid.New()
	msg.Kind = format.ServiceStart
	msg.Data = msgUUID
	
	err := enc.Encode(msg)
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}

	msg.Kind = format.ServiceFinish
	err = enc.Encode(msg)
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
}


