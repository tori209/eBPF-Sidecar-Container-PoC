package main

import (
	"io"
	"log"
	"os"
	"net"
	"fmt"
	"encoding/gob"
	"strings"
	"context"

  	"github.com/tori209/data-executor/log/format"
  	"github.com/tori209/data-executor/log/db_access"
)

func initInfluxdbOption() (*db_access.InfluxdbOptions) {
	org := os.Getenv("COLLECTOR_ORG")
	if org == "" {
		return nil
	}
	token := strings.TrimSpace(os.Getenv("SECRET_COLLECTOR_TOKEN"))
	if token == "" {
		return nil
	}
	bucket := os.Getenv("COLLECTOR_BUCKET")
	if bucket == "" {
		return nil
	}
	url := os.Getenv("INFLUXDB_URL")
	if url == "" {
		return nil
	}
	return &db_access.InfluxdbOptions{
		Url: url,
		Bucket: bucket,
		Org: org,
		Token: token,
	}
}

func initPostgresOption() (*db_access.PostgresDbOpt) {
	postgresDSN := os.Getenv("POSTGRES_TASKDB_DSN")
	if postgresDSN == "" {
		log.Fatalf("POSTGRES_TASKDB_DSN not found in env.")
	}

	return &db_access.PostgresDbOpt{
		DSN: postgresDSN,
		Ctx: context.Background(),
	}

}

func main() {
	// Env 확인 ======================================================
	if os.Getenv("COLLECTOR_ENV_READY") != "Ready" {
		log.Fatalf("Environment is not set. Apply ConfigMap in config/executor-pod.yaml.")
	}

	socketPath := os.Getenv("COLLECTOR_SOCK_PATH")
	if socketPath == "" {
		log.Fatalf("COLLECTOR_SOCK_PATH not found in env")
	}

	// Socket Check ==================================================
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}

	// DB Connection Check ===========================================
	//opt := *(initInfluxdbOption())
	//iqr := db_access.NewInfluxdbQueryRunner(opt)
	pqr := db_access.NewPostgresQueryRunner(
		initPostgresOption(),
	)

	// Ready Sequence Done. ==========================================
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
		go receive_message(conn, pqr)
	}
}

func receive_message(conn net.Conn, lqr db_access.LogQueryRunner) {
	defer conn.Close()

	dec := gob.NewDecoder(conn)
	var msgSlice []format.L4Message
	for {
		err := dec.Decode(&msgSlice)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Read Failed: %+v\n", err)
			break
		}
		
		go func() {
			go_err := lqr.SaveLogs(&msgSlice)
			if go_err != nil {
				log.Printf("DB Save Failed: %+v", go_err)
			}
		}()

		for _, msg := range msgSlice {
			fmt.Printf(msg.String() + "\n")
		}
	}
}
