package report

import (
	"net"
	"encoding/gob"
	"log"
	
	"github.com/google/uuid"
	"github.com/tori209/data-executor/log/format"
)

// In Executor, Runner->Watcher Report
type Reporter struct {
	proto		string
	targetPath	string
}

//============================================================

func NewReporter(protocol, targetPath string) (*Reporter) {
	return &Reporter{
		proto: protocol,
		targetPath: socketPath,
	}
}

//============================================================

func (r *Reporter) createConnection() (*net.Conn, error) {
	conn, err := net.Dial(r.proto, r.targetPath)
	if err != nil {
		log.Printf("[Reporter.createConnection] Failed to create Connection")
		return nil, err
	}
	return &conn, err
}

//============================================================

func (r *Reporter) ReportTaskStart(JobID, TaskID uuid.UUID) error {
	conn, err := r.createConnection()
	if err != nil {
		log.Printf("[Reporter.ReportTaskStart] Failed to create Connection %+v:", err)
		return err
	}
	defer conn.Close()
	reportEnc := gob.NewEncoder(conn)

	// Message Send
	err := reportEnc.Encode(format.ReportMessage{
		Kind: format.TaskStart,
		JobID: JobID,
		TaskID: TaskID,
	})
	if err != nil {
		log.Printf("[Reporter.ReportTaskStart] Report Message Send Failed (JobID: %s, TaskID: %s) %+v:",
			r.jobID.String(),
			r.taskID.String(),
			err,
		)
	} else {
		log.Printf("[Reporter.ReportTaskStart] Report Task Start (JobID: %s, TaskID: %s)",
			r.jobID.String(),
			r.taskID.String(),
		)
	}

	return err
}

//============================================================

func (r *Reporter) ReportTaskResult(success bool, JobID, TaskID uuid.UUID) error {
	conn, err := r.createConnection()
	if err != nil {
		log.Printf("[Reporter.ReportTaskResult] Failed to create Connection %+v:", err)
		return err
	}
	defer conn.Close()
	reportEnc := gob.NewEncoder(conn)

	var reportKind format.ReportType
	if success {
		reportKind = format.TaskFinish
	} else {
		reportKind = format.TaskFailed
	}

	err := reportEnc.Encode(format.ReportMessage{
		Kind: reportKind,
		JobID: r.jobID,
		TaskID: r.taskID,
	})
	if err != nil {
		log.Printf("[Reporter.ReportTaskResult] Report Message Send Failed %+v:", err)
	}

	return err
}

//============================================================

func (r *Reporter) ReportRunnerStart() error {
	conn, err := r.createConnection()
	if err != nil {
		log.Printf("[Reporter.ReportRunnerStart] Failed to create Connection %+v:", err)
		return err
	}
	defer conn.Close()
	reportEnc := gob.NewEncoder(conn)


	if err := reportEnc.Encode(format.ReportMessage{
		Kind: format.RunnerStart,
	}); err != nil {
		log.Printf("[Reporter.ReportRunnerStart] Report Message Send Failed: %+v", err)
		return err
	}
	return nil
}

//============================================================

func (r *Reporter) ReportRunnerFinish() error {
	conn, err := r.createConnection()
	if err != nil {
		log.Printf("[Reporter.ReportRunnerFinish] Failed to create Connection %+v:", err)
		return err
	}
	defer conn.Close()
	reportEnc := gob.NewEncoder(conn)


	if err := r.reportEnc.Encode(format.ReportMessage{
		Kind: format.RunnerFinish,
	}); err != nil {
		log.Printf("[Reporter.ReportRunnerFinish] Report Message Send Failed: %+v", err)
		return err
	}
	return nil
}
