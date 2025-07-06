package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/tori209/data-executor/log/report"
	"github.com/tori209/data-executor/runner/coderunner"
	//"github.com/tori209/data-executor/log/format"
)

func initDestinationConfig() *coderunner.DestinationConfig {
	normalEndpoint := os.Getenv("NORMAL_MINIO_ENDPOINT")
	if normalEndpoint == "" {
		log.Fatalf("[Runner] NORMAL_MINIO_ENDPOINT not defined.")
	}
	normalBucket := os.Getenv("NORMAL_MINIO_BUCKET")
	if normalBucket == "" {
		log.Fatalf("[Runner] NORMAL_MINIO_BUCKET not defined.")
	}
	malEndpoint := os.Getenv("MALICIOUS_MINIO_ENDPOINT")
	if malEndpoint == "" {
		log.Fatalf("[Runner] MALICIOUS_MINIO_ENDPOINT not defined.")
	}
	malBucket := os.Getenv("MALICIOUS_MINIO_BUCKET")
	if malBucket == "" {
		log.Fatalf("[Runner] MALICIOUS_MINIO_BUCKET not defined.")
	}
	minioID := os.Getenv("MINIO_COMMON_ID")
	if minioID == "" {
		log.Fatalf("[Runner] MINIO_COMMON_ID not defined.")
	}
	minioPW := os.Getenv("MINIO_COMMON_PW")
	if minioPW == "" {
		log.Fatalf("[Runner] MINIO_COMMON_PW not defined.")
	}

	return &coderunner.DestinationConfig{
		NormalCaseEndpoint:   normalEndpoint,
		NormalCaseBucket:     normalBucket,
		AbnormalCaseEndpoint: malEndpoint,
		AbnormalCaseBucket:   malBucket,
		MinioID:              minioID,
		MinioPW:              minioPW,
	}
}

func main() {
	// 환경변수 설정 유무 확인
	socketPath := os.Getenv("WATCHER_SOCK_PATH")
	if socketPath == "" {
		log.Fatalf("[Runner] WATCHER_SOCK_PATH not defined.")
	}
	socketType := os.Getenv("WATCHER_SOCK_TYPE")
	if socketType == "" {
		log.Fatalf("[Runner] WATCHER_SOCK_TYPE not defined.")
	}

	driverContactProto := os.Getenv("DRIVER_CONTACT_PROTO")
	if driverContactProto == "" {
		log.Fatalf("[Runner] DRIVER_CONTACT_PROTO not defined.")
	}
	driverContactFQDN := os.Getenv("DRIVER_CONTACT_FQDN") // 임시로 환경변수로 설정. 다른 방법이 안떠오르네.
	if driverContactFQDN == "" {
		log.Fatalf("[Runner] DRIVER_CONTACT_FQDN not defined.")
	}

	runnerRequestProto := os.Getenv("RUNNER_REQUEST_RECEIVE_PROTO")
	if runnerRequestProto == "" {
		log.Fatalf("[Runner] RUNNER_REQUEST_RECEIVE_PROTO not defined.")
	}
	runnerRequestPort := os.Getenv("RUNNER_REQUEST_RECEIVE_PORT")
	if runnerRequestPort == "" {
		log.Fatalf("[Runner] RUNNER_REQUEST_RECEIVE_PORT not defined.")
	}

	// Watcher 연결 시도
	log.Printf("[Runner] Try to Report to Watcher...")
	watcherReporter := report.NewWatcherReporter(socketType, socketPath)
	for {
		if err := watcherReporter.ReportRunnerStart(); err != nil {
			log.Printf("[Runner] Failed to send Report. Wait...: %+v", err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	defer watcherReporter.ReportRunnerFinish()
	log.Printf("[Runner] Report sent to Watcher.")

	// Driver와의 연결 시도.
	log.Printf("[Runner] Try to Report to Driver...")
	driverReporter := report.NewReporter(driverContactProto, driverContactFQDN)
	for {
		if err := driverReporter.ReportRunnerStart(); err != nil {
			log.Printf("[Runner] Failed to send Report. Wait...: %+v", err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	defer driverReporter.ReportRunnerFinish()
	log.Printf("[Runner] Report sent to Driver.")

	// 초기화 완료 ===========================================================

	// 초기화 후 Task 처리 영역
	// 한 순간에 하나의 Task만 처리한다고 가정.

	listener, err := net.Listen(runnerRequestProto, runnerRequestPort)
	if err != nil {
		log.Fatalf("[Runner] Failed to create Request Listener: %+v", err)
	}

	rpc.Register(coderunner.NewCodeRunner(watcherReporter, initDestinationConfig()))
	log.Printf("[Runner] Waiting for new task...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[Runner] Connection Failed: %+v", err)
			continue
		}
		rpc.ServeConn(conn)
	}
}
