package main

import (
	"io"
	"log"
	"os"
	"net"
	"fmt"
	"encoding/gob"

  	"github.com/tori209/data-executor/log/format"
)


func main() {
	if os.Getenv("COLLECTOR_ENV_READY") != "Ready" {
		log.Fatalf("Environment is not set. Apply ConfigMap in config/executor-pod.yaml.")
	}

	socketPath := os.Getenv("COLLECTOR_SOCK_PATH")
	if socketPath == "" {
		log.Fatalf("COLLECTOR_SOCK_PATH not found in env")
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
		go receive_message(conn)
	}
}

func receive_message(conn net.Conn) {
	defer conn.Close()

	dec := gob.NewDecoder(conn)
	var msg format.L4Message
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
		fmt.Printf(msg.String() + "\n");
	}
}
