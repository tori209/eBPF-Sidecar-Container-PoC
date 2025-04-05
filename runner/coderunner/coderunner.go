package coderunner

import (
	"time"

	"github.com/tori209/data-executor/log/format"
	"github.com/tori209/data-executor/log/report"
)

type CodeRunner struct {
	watcherReporter		*report.Reporter
}

func NewCodeRunner(reporter *report.Reporter) (*CodeRunner) {
	return &CodeRunner{
		watcherReporter: reporter
	}
}

func (cr *CodeRunner) preRunningProcedure(reqMsg *format.TaskRequestMessage) {
	var i int
	for i = 0; i < 5; i++ {
		if err := reporter.ReportTaskStart(reqMsg.JobID, reqMsg.TaskID); err != nil {
			log.Printf("[Runner] Reporter Failed to send runner start. Wait...: %+v",err)
			time.Sleep(3 * time.Second)
		} else {  break  }
	}
	if (i == 5) {
		log.Fatalf("[Runner] Reporter Failed to send to watcher. Seems like wathcer died. Shutdown...")
	}
}

func (cr *CodeRunner) postRunningProcedure(resMsg *format.TaskResponseMessage, result bool) {
	// Report Task Done to Watcher
	var i int
	for i = 0; i < 5; i++ {
		if err := reporter.ReportTaskResult(result); err != nil {
			log.Printf("[Runner] Reporter Failed to send task result. Wait...: %+v",err)
			time.Sleep(3 * time.Second)
		} else {  break;  }
	}
	if (i == 5) {
		log.Fatalf("[Runner] Reporter Failed to send to watcher. Seems like wathcer died. Shutdown...")
	}
}

// Dummy Task Request
func (cr *CodeRunner) StartTask(reqMsg *format.TaskRequestMessage, resMsg *format.TaskResponseMessage) (error) {
	cr.preRunningProcedure(reqMsg)

	time.Sleep(5 * time.Second)

	resMsg.JobID = reqMsg.JobID
	resMsg.TaskID = reqMsg.TaskID
	resMsg.Status = format.TaskFinish
	cr.postRunningProcedure(reqMsg)
	return nil
}
