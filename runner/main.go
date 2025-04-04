package main

import (
    "encoding/gob"
	"log"
    "net"
    "time"
	"os"

	"github.com/tori209/data-executor/runner/coderunner"
	"github.com/tori209/data-executor/runner/report"
	"github.com/tori209/data-executor/log/format"
)

func main() {
	// 환경변수 설정 유무 확인
	socketPath := os.Getenv("WATCHER_SOCK_PATH")
	if socketPath == "" {
		log.Fatalf("[Runner] WATCHER_SOCK_PATH not defined.")
	}
	driverContactFQDN := os.Getenv("DRIVER_CONTACT_FQDN") // 임시로 환경변수로 설정. 다른 방법이 안떠오르네.
	if driverContactFQDN == "" {
		log.Fatalf("[Runner] DRIVER_CONTACT_FQDN not defined.")
	}


	log.Printf("[Runner] Try to Connect to Watcher...")
	var watcherConn net.Conn
	for {
		con, err := net.Dial("unix", socketPath)
		if err == nil {
			watcherConn = con
			break
		}
		log.Printf("[Runner] Failed to connect socket %s: %+v", socketPath, err)
		time.Sleep(1 * time.Second)
	}
	defer watcherConn.Close()
	log.Printf("[Runner] Connection to Watcher established.")

	log.Printf("[Runner] Try to Connect to Driver...")
	var reqListener net.Listener
	for {
		listener, err := net.Listen("tcp", "0.0.0.0:8080")
		if err == nil {
			reqListener = listener
			break
		}
		log.Printf("[Runner] Failed to create listening socket: %+v", err)
		time.Sleep(1 * time.Second)
	}
	defer reqListener.Close()
	log.Printf("[Runner] Connection to Driver established.")

	reporter := report.NewReporter(watcherConn)
	if err := reporter.ReportRunnerStart(); err != nil {
		log.Fatalf("[Runner] Failed to send Report: %+v", err)
	}
	defer reporter.ReportRunnerFinish()

	log.Printf("[Runner] Waiting for new task...")
	// 초기화 후 Task 처리 영역
	// 한 순간에 하나의 Task만 처리한다고 가정.
	for {
		tcpConn, err := reqListener.Accept()
		if err != nil {
			log.Printf("[Runner] Failed to receive task request: %+v", err)
			continue
		}
		log.Printf("[Runner] Received New Task!")
		reqDec := gob.NewDecoder(tcpConn)
		resEnc := gob.NewEncoder(tcpConn)

		var reqMsg format.TaskRequestMessage
		var resMsg format.TaskResponseMessage

		// Request 수신 시도, 실패 시 수신 실패 안내 후 Close
		if err := reqDec.Decode(&reqMsg); err != nil {
			log.Printf("[Runner] TaskRequest Decode failed: %+v", err)
			resMsg.Status = format.MessageBroken
			resEnc.Encode(resMsg)
			tcpConn.Close()
			continue
		}
		
		cr := coderunner.NewCodeRunner(reqMsg) 

		// Report Task Start to Watcher
		var i int
		for i = 0; i < 5; i++ {
			if err := reporter.ReportTaskStart(reqMsg.JobID, reqMsg.TaskID); err == nil {
				break
			}
			log.Printf("[Runner] Reporter Failed to send runner start. Wait...: %+v",err)
			time.Sleep(3 * time.Second)
		}
		if (i == 5) {
			log.Fatalf("[Runner] Reporter Failed to send to watcher. Seems like wathcer died. Shutdown...")
		}


		// Start Task (Block Exec)
		log.Printf("[Runner] Now start running code...")
		result, err := cr.StartTask()
		if err != nil {
			log.Printf("[Runner] Error Occured During task: %+v", err)
		}
		log.Printf("[Runner] Task Complete. Now try to report to Watcher...")

		// Report Task Done to Watcher
		for i = 0; i < 5; i++ {
			if err := reporter.ReportTaskResult(result); err != nil {
				log.Printf("[Runner] Reporter Failed to send task result. Wait...: %+v",err)
			}
		}
		if (i == 5) {
			log.Fatalf("[Runner] Reporter Failed to send to watcher. Seems like wathcer died. Shutdown...")
		}

		log.Printf("[Runner] Reporting Complete. Now try to response to Driver...")
		// Send TaskResponse to Manager
		resMsg.JobID = reqMsg.JobID
		resMsg.TaskID = reqMsg.TaskID
		if result {
			resMsg.Status = format.TaskFinish
		} else {
			resMsg.Status = format.TaskFailed
		}
		if err := resEnc.Encode(resMsg); err != nil {
			log.Printf("[Runner] TaskResponse Encode failed: %+v", err)
		} else {
			log.Printf("[Runner] Task Done.")
		}
		tcpConn.Close()
	}
}
