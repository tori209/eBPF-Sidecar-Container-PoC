package report

import (
	"net/rpc"
	"log"
	
	"github.com/google/uuid"
	"github.com/tori209/data-executor/log/format"
)

// In Executor, Runner->Watcher Report
type WatcherReporter struct {
	proto	string
	targetPath	string
}

//============================================================

func NewWatcherReporter(protocol, targetPath string) (*WatcherReporter) {
	return &WatcherReporter{
		proto: protocol,
		targetPath: targetPath,
	}
}

//============================================================

func (r *WatcherReporter) createRpcClient() (*rpc.Client, error) {
	client, err := rpc.Dial(r.proto, r.targetPath)
	if err != nil {
		log.Printf("[WatcherReporter.createRpcClient] Failed to create RPC Client")
		return nil, err
	}
	return client, err
}

//============================================================

func (r *WatcherReporter) ReportTaskStart(jobID, taskID uuid.UUID) error {
	client, err := r.createRpcClient()
	if err != nil {
		log.Printf("[WatcherReporter.ReportTaskStart] Failed to create Connection %+v:", err)
		return err
	}
	defer client.Close()

	// Report 전송
	request := format.ReportMessage{
		Kind: format.TaskStart,
		JobID: jobID,
		TaskID: taskID,
	}
	var response format.ReportMessage
	err = client.Call("BpfTrafficCapture.ReceiveTaskMessage", &request, &response)

	if err != nil {
		log.Printf("[WatcherReporter.ReportTaskStart] Report Message Send Failed (JobID: %s, TaskID: %s) %+v:",
			jobID.String(),
			taskID.String(),
			err,
		)
	} else {
		log.Printf("[WatcherReporter.ReportTaskStart] Report Task Start (JobID: %s, TaskID: %s)",
			jobID.String(),
			taskID.String(),
		)
	}

	return err
}

//============================================================

func (r *WatcherReporter) ReportTaskResult(success bool, jobID, taskID uuid.UUID) error {
	client, err := r.createRpcClient()
	if err != nil {
		log.Printf("[WatcherReporter.ReportTaskResult] Failed to create Connection %+v:", err)
		return err
	}
	defer client.Close()

	// Report 전송
	request := format.ReportMessage{
		Kind: format.TaskStart,
		JobID: jobID,
		TaskID: taskID,
	}
	if success {
		request.Kind = format.TaskFinish
	} else {
		request.Kind = format.TaskFailed
	}
	var response format.ReportMessage
	err = client.Call("BpfTrafficCapture.ReceiveTaskMessage", &request, &response)
	
	if err != nil {
		log.Printf("[WatcherReporter.ReportTaskResult] Report Message Send Failed %+v:", err)
	}

	return err
}

//============================================================

func (r *WatcherReporter) ReportRunnerStart() error {
	client, err := r.createRpcClient()
	if err != nil {
		log.Printf("[WatcherReporter.ReportRunnerStart] Failed to create Connection %+v:", err)
		return err
	}
	defer client.Close()

	request := format.ReportMessage{
		Kind: format.RunnerStart,
	}
	var response format.ReportMessage
	err = client.Call("BpfTrafficCapture.ReceiveTaskMessage", &request, &response)

	if err != nil {
		log.Printf("[WatcherReporter.ReportRunnerStart] Report Message Send Failed: %+v", err)
		return err
	}
	return nil
}

//============================================================

func (r *WatcherReporter) ReportRunnerFinish() error {
	client, err := r.createRpcClient()
	if err != nil {
		log.Printf("[WatcherReporter.ReportRunnerFinish] Failed to create Connection %+v:", err)
		return err
	}
	defer client.Close()

	request := format.ReportMessage{
		Kind: format.RunnerFinish,
	}
	var response format.ReportMessage
	err = client.Call("BpfTrafficCapture.ReceiveTaskMessage", &request, &response)

	if err != nil {
		log.Printf("[WatcherReporter.ReportRunnerFinish] Report Message Send Failed: %+v", err)
		return err
	}
	return nil
}
