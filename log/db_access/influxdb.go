package db_access

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tori209/data-executor/log/format"
	
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	// "github.com/influxdata/influxdb-client-go/v2/api/write"
	apiHttp "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxdbOptions struct {
	Url	string
	Bucket string
	Org	string
	Token string
}

type InfluxdbQueryRunner struct {
	client		influxdb2.Client
	writeAPI	api.WriteAPI
}

func NewInfluxdbQueryRunner (opt InfluxdbOptions) (*InfluxdbQueryRunner) {
	client := influxdb2.NewClient(opt.Url, opt.Token)
	writeApi := client.WriteAPI(opt.Org, opt.Bucket)

	iqr := &InfluxdbQueryRunner {
		client: client,
		writeAPI: writeApi,
	}

	go func() {
		for err := range writeApi.Errors() {
			log.Printf("[InfluxdbQueryRunner] Failed to write (trace-id: %s): %+v", err.(*apiHttp.Error).Header.Get("Trace-ID"), err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		log.Printf("Termination Signal Received. ")
		iqr.Close()
		os.Exit(0)
	}()

	return iqr
}

func (iqr * InfluxdbQueryRunner) InsertLog(msg format.L4Message) {
	p := influxdb2.NewPoint(
		"task_l4_traffic",
		map[string]string{
			"job_id": msg.JobID.String(),
		},
		map[string]interface{}{
			"task_id": msg.TaskID,
			"src": msg.GetSrcAsString(),
			"dst": msg.GetDstAsString(),
			"proto": format.GetL4ProtoString(msg.L4Proto),
			"size": msg.Size,
		},
		time.Unix(0, msg.TS),
	)
	iqr.writeAPI.WritePoint(p)
}

func (iqr *InfluxdbQueryRunner) Close() {
	iqr.writeAPI.Flush()
	iqr.client.Close()
}
