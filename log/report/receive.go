package report

import (
	"net"
	"encoding/gob"
	"log"

	"github.com/google/uuid"
	"github.com/tori209/data-executor/log/format"
)

type Receiver struct {
	proto	string
	address	string
}

//============================================================

func NewReceiver(proto, address string) (*Receiver) {
	return &Receiver{
		proto: proto,
		address: address
	}
}

