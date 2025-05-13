package main

import (
	"log"
	"os"
	"time"
	"strconv"
	"context"
	"math/rand"
	
	"github.com/google/uuid"
	"github.com/tori209/data-executor/driver/manage"
	"github.com/tori209/data-executor/log/format"
	"github.com/tori209/data-executor/log/db_access"
)

func main() {
	// Check env =====================
	str_num := os.Getenv("DESIRED_EXECUTOR_NUMBER")
	if str_num == "" {
		log.Fatalf("DESIRED_EXECUTOR_NUMBER not found in env.")
	}
	desiredExecutorNum, conv_err := strconv.Atoi(str_num)
	if conv_err != nil {
		log.Fatalf("DESIRED_EXECUTOR_NUMBER is not a number..")
	}
	postgresDSN := os.Getenv("POSTGRES_TASKDB_DSN")
	if postgresDSN == "" {
		log.Fatalf("POSTGRES_TASKDB_DSN not found in env.")
	}

	// Create DB QueryRunner =========
	pqr := db_access.NewPostgresQueryRunner(
		&db_access.PostgresDbOpt{
			DSN: postgresDSN,
			Ctx: context.Background(),
		},
	)
	if err := pqr.Init(); err != nil {
		log.Fatalf("[Driver/main] DB Init Failed: %+v", err)
	}
	log.Printf("[Driver/main] PostgresSQL Connection Established.")
	
	// Create Manager ================
	var em *manage.ExecutorManager
	if manager, err := manage.NewExecutorManager(":8080", pqr); err != nil {
		log.Fatalf("[Driver/main] Failed to create manager: %+v", err)
	} else {
		em = manager
	}

	// 요청 받는 것만 확인.
	for {
		cnt := em.CountOnlineExecutor() 
		if cnt >= desiredExecutorNum {  break  }
		log.Printf("[Driver/main] Waiting executors being online... Current: %d", cnt)
		time.Sleep(time.Second)
	}
	log.Printf("[Driver/main] Desired Number fulfilled.")

	em.ProcessJob(
		format.TaskRequestMessage{
			JobID: uuid.New(),
			TaskID: uuid.New(),
			DataSource: format.DataSourceInfo{
				Endpoint: "minio.minio-s.svc.cluster.local:80",
				BucketName: "dummy-bucket",
				ObjectName: "dummy_sensor_data.csv",
			},
			DestinationURL: "",
			RangeBegin:	int64(0),
			RangeEnd:	int64(1000),
			RunAsEvil: rand.Intn(2) == 0,
		},
		int64(100),
	)

	em.Destroy()
	log.Printf("[Driver/main] Job Finished.")
}
