package frontdesk

import (
	"net"
	"net/http"
	"time"
	"sync"
)

const MAX_CONN_RETRY=100
const DEFAULT_RECEIVER_PORT=":8080"

type FrontdeskOption struct {
	reportSockType	string,
	reportSockPath	string,
}

type Frontdesk struct {
	reportConn		*net.Conn,
	receiver		*http.Server
	mu				*
}

func NewFrontdesk(option FrontdeskOption) (*Frontdesk) {
	for i := 0; i < MAX_CONN_RETRY; i++ {
		if con, err := net.Dial(option.reportSockType, option.reportSockPath); err == nil {
			// connection established
			return new(Frontdesk{
				reportConn: con,
				receiver:	nil,
				mu:			new(sync.Mutex),
			})
		}
	}
	return nil
}

func (fd *Frontdesk) StartWork() {
	fd.receiver = &http.Server{
		Addr:	DEFAULT_RECEIVER_PORT,
		Handler:	nil,	// DefaultServeMux
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := fd.receiver.ListenAndServe()
	log.Printf("Frontdesk Receive Server Closed with ErrCode: %+v", err)
}

func (fd *Frontdesk) StopWork() {
	fd.receiver.Shutdown()
}

func (fd *Frontdesk) Destory() {
	fd.receiver.Close()
	fd.reportConn.Close()
}

