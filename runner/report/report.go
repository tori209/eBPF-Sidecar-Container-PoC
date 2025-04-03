package report

import (
	"net"
	"encoding/gob"
	"log"
//	"time"
	"errors"
	
	"github.com/tori209/data-executor/log/format"
)

//const MAX_REPORT_RETRY = 100
//const REPORT_RETRY_INTERVAL = 1 * time.Seconds

// In Executor, Runner->Excutor Report
type Reporter struct {
	reportEnc	*gob.Encoder
	jobID		[16]byte
	taskID		[16]byte
	taskRunning	bool
}

func NewReporter(conn net.Conn) (*Reporter) {
	return &Reporter{
		reportEnc: gob.NewEncoder(conn),
	}
}

func (r *Reporter) ReportTaskStart(JobID, TaskID [16]byte) error {
	if r.taskRunning {  
		return errors.New("Task Not yet Finished")
	}

	err := r.reportEnc.Encode(format.ReportMessage{
		Kind: format.TaskStart,
		JobID: JobID,
		TaskID: TaskID,
	})
	if err != nil {
		log.Printf("[Reporter.ReportTaskStart] Report Message Send Failed %+v:", err)
	}

	r.jobID = JobID
	r.taskID = TaskID
	r.taskRunning = true

	return err
}

func (r *Reporter) ReportTaskResult(success bool) error {
	if !r.taskRunning {  
		return errors.New("Task Not yet Finished")
	}

	var reportKind format.ReportType
	if success {
		reportKind = format.TaskFinish
	} else {
		reportKind = format.TaskFailed
	}

	err := r.reportEnc.Encode(format.ReportMessage{
		Kind: reportKind,
		JobID: r.jobID,
		TaskID: r.taskID,
	})
	if err != nil {
		log.Printf("[Reporter.ReportTaskResult] Report Message Send Failed %+v:", err)
	}

	r.taskRunning = false
	r.jobID = [16]byte{}  // Clear
	r.taskID = [16]byte{}

	return err
}

func (r *Reporter) ReportRunnerStart() error {
	if err := r.reportEnc.Encode(format.ReportMessage{
		Kind: format.RunnerStart,
	}); err != nil {
		log.Printf("[Reporter.ReportRunnerStart] Report Message Send Failed: %+v", err)
		return err
	}
	return nil
}

func (r *Reporter) ReportRunnerFinish() error {
	if err := r.reportEnc.Encode(format.ReportMessage{
		Kind: format.RunnerFinish,
	}); err != nil {
		log.Printf("[Reporter.ReportRunnerStart] Report Message Send Failed: %+v", err)
		return err
	}
	return nil
}
