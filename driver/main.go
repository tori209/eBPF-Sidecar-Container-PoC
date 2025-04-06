package main

import (
	"log"
	"os"
	"time"
	"strconv"
	
	"github.com/google/uuid"
	"github.com/tori209/data-executor/driver/manage"
	"github.com/tori209/data-executor/log/format"
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
	
	// Create Manager ================
	var em *manage.ExecutorManager
	if manager, err := manage.NewExecutorManager(":8080"); err != nil {
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
			DataSourceURL: "",
			DestinationURL: "",
			RangeBegin:	uint64(0),
			RangeEnd:	uint64(1000),
			RunAsEvil: false,
		},
		uint64(100),
	)

	em.Destroy()
	log.Printf("[Driver/main] Job Finished.")
}
