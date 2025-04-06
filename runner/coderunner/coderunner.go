package coderunner

import (
	"time"
	"log"

	"github.com/tori209/data-executor/log/format"
	"github.com/tori209/data-executor/log/report"
)

type CodeRunner struct {
	watcherReporter		*report.WatcherReporter
}

func NewCodeRunner(reporter *report.WatcherReporter) (*CodeRunner) {
	return &CodeRunner{
		watcherReporter: reporter,
	}
}

func (cr *CodeRunner) preRunningProcedure(reqMsg *format.TaskRequestMessage) {
	log.Printf("[Runner] Start Task (Job: %s / Task: %s)", reqMsg.JobID.String(), reqMsg.TaskID.String())
	var i int
	for i = 0; i < 5; i++ {
		if err := cr.watcherReporter.ReportTaskStart(reqMsg.JobID, reqMsg.TaskID); err != nil {
			log.Printf("[Runner] Reporter Failed to send runner start. Wait...: %+v",err)
			time.Sleep(3 * time.Second)
		} else {  break  }
	}
	if (i == 5) {
		log.Fatalf("[Runner] Reporter Failed to send to watcher. Seems like wathcer died. Shutdown...")
	}
}

func (cr *CodeRunner) postRunningProcedure(resMsg *format.TaskResponseMessage) {
	// Report Task Done to Watcher
	var isSuccess bool
	if resMsg.Status == format.TaskFinish {  
		isSuccess = true
	} else {  isSuccess = false  }

	var i int
	for i = 0; i < 5; i++ {
		if err := cr.watcherReporter.ReportTaskResult(isSuccess, resMsg.JobID, resMsg.TaskID); err != nil {
			log.Printf("[Runner] Reporter Failed to send task result. Wait...: %+v",err)
			time.Sleep(3 * time.Second)
		} else {  break;  }
	}
	if (i == 5) {
		log.Fatalf("[Runner] Reporter Failed to send to watcher. Seems like wathcer died. Shutdown...")
	}
	log.Printf("[Runner] Finish Task (Job: %s / Task: %s)", resMsg.JobID.String(), resMsg.TaskID.String())
}

// Dummy Task Request
func (cr *CodeRunner) StartTask(reqMsg *format.TaskRequestMessage, resMsg *format.TaskResponseMessage) (error) {
	cr.preRunningProcedure(reqMsg)

	time.Sleep(5 * time.Second)

	resMsg.JobID = reqMsg.JobID
	resMsg.TaskID = reqMsg.TaskID
	resMsg.Status = format.TaskFinish
	cr.postRunningProcedure(resMsg)
	return nil
}
