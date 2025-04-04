package main

import (
	"log"
	"os"
	"time"
	"strconv"
	
	"github.com/tori209/data-executor/driver/manage"
)

func main() {
	// Check env =====================
	controlPort := os.Getenv("DRIVER_CONTROL_PORT")
	if controlPort == "" {
		log.Fatalf("DRIVER_CONTROL_PORT not found in env.")
	}
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
	if manager, err := manage.NewExecutorManager(controlPort); err != nil {
		log.Fatalf("[Driver/main] Failed to create manager: %+v", err)
	} else {
		em = manager
	}

	for {
		cnt := em.CountOnlineExecutor() 
		if cnt >= desiredExecutorNum {  break  }
		log.Printf("[Driver/main] Waiting executors being online... Current: %d", cnt)
		time.Sleep(time.Second)
	}

	em.Destroy()
}
