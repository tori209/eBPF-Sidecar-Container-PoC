package coderunner

import (
	"time"

	"github.com/tori209/data-executor/log/format"
)

type CodeRunner struct {
	request		format.TaskRequestMessage
	start		time.Time
}

func startDummyScript() {
	time.Sleep(time.Second)
}

func NewCodeRunner(reqMessage format.TaskRequestMessage) (*CodeRunner) {
	return &CodeRunner{
		request:	reqMessage,
		start:		time.Now(),
	}
}

// Dummy Task Request
func (cr *CodeRunner) StartTask() (bool, error) {
	time.Sleep(5 * time.Second)
	return true, nil
}
