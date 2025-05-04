package coderunner

import (
	"fmt"
	"bytes"
	"time"
	"log"
	"context"
	"io"
	"compress/gzip"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/tori209/data-executor/log/format"
	"github.com/tori209/data-executor/log/report"
)

type DestinationConfig struct {
	NormalCaseEndpoint		string
	NormalCaseBucket		string
	AbnormalCaseEndpoint	string
	AbnormalCaseBucket		string
	MinioID					string
	MinioPW					string
}

type CodeRunner struct {
	watcherReporter		*report.WatcherReporter
	dstConf				*DestinationConfig
}

func NewCodeRunner(reporter *report.WatcherReporter, dstConf *DestinationConfig) (*CodeRunner) {
	return &CodeRunner{
		watcherReporter: reporter,
		dstConf: dstConf,
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

// StartTask의 오류 발생 시 수행할 역할.
func (cr *CodeRunner) errorProcessing(formatString string, err error, resMsg *format.TaskResponseMessage) {
	log.Printf(formatString, err)
	cr.postRunningProcedure(resMsg)
}

// Task Request. Driver에서 Executor로 Go RPC 라이브러리를 통해 실행.
func (cr *CodeRunner) StartTask(reqMsg *format.TaskRequestMessage, resMsg *format.TaskResponseMessage) (error) {
	cr.preRunningProcedure(reqMsg)

	taskFailedProc := func () {
		resMsg.JobID = reqMsg.JobID
		resMsg.TaskID = reqMsg.TaskID
		resMsg.Status = format.TaskFailed
	}

	s3Client, err := minio.New(reqMsg.DataSource.Endpoint, &minio.Options{
		Secure: false,
	})
	if err != nil {
		taskFailedProc()
		cr.errorProcessing("[Runner] Failed establish connection to DataSource: %+v", err, resMsg)
		return err
	}

	// 1. GetObject 수행
	getOpt := minio.GetObjectOptions{}
	getOpt.SetRange(reqMsg.RangeBegin, reqMsg.RangeEnd)
	reader, err := s3Client.GetObject(
			context.Background(),
			reqMsg.DataSource.BucketName,
			reqMsg.DataSource.ObjectName,
			getOpt,
	)
	if err != nil {
		taskFailedProc()
		cr.errorProcessing("[Runner] Failed to execute 'GetObject': %+v", err, resMsg)
		return err
	}
	defer reader.Close()

	// 2. Data 가져오기 시도
	buf := new(bytes.Buffer)
	if _, err  := io.Copy(buf, reader); err != nil {
		taskFailedProc()
		cr.errorProcessing("[Runner] Failed to create buffer: %+v", err, resMsg)
		return err
	}
	
	compressedData, err := compressGzip(buf.Bytes())
	if err != nil {
		taskFailedProc()
		cr.errorProcessing("[Runner] Failed to compress: %+v", err, resMsg)
		return err
	}

	// 3. 전송 경로 다분화
	var uploadEndpoint, uploadBucket, uploadObject string
	if reqMsg.RunAsEvil {
		uploadEndpoint = cr.dstConf.AbnormalCaseEndpoint
		uploadBucket = cr.dstConf.AbnormalCaseBucket
	} else {
		uploadEndpoint = cr.dstConf.NormalCaseEndpoint
		uploadBucket = cr.dstConf.NormalCaseBucket
	}
	uploadObject = fmt.Sprintf("%s.%s.%d_to_%d.gzip", 
		reqMsg.JobID.String(), reqMsg.DataSource.ObjectName,
		reqMsg.RangeBegin, reqMsg.RangeEnd,
	)

	// 4. 저장 시도
	dstClient, err := minio.New(uploadEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cr.dstConf.MinioID, cr.dstConf.MinioPW, ""),
		Secure: false,
	})
	dstReader := bytes.NewReader(compressedData)

	uploadInfo, err := dstClient.PutObject(
		context.Background(), uploadBucket, uploadObject, dstReader, 
		int64(dstReader.Len()), 
		minio.PutObjectOptions{ContentType: "application/gzip"},
	)
	if err != nil {
		taskFailedProc()
		cr.errorProcessing("[Runner] Failed to upload data: %+v", err, resMsg)
		return err
	}
	log.Printf("[Runner] Upload Success: %s of size %d", uploadInfo.Key, uploadInfo.Size)
	// 작업 완료 보고
	resMsg.JobID = reqMsg.JobID
	resMsg.TaskID = reqMsg.TaskID
	resMsg.Status = format.TaskFinish
	cr.postRunningProcedure(resMsg)
	return nil
}

func compressGzip(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

