package manage

import (
	"os"
	"net"
	"net/rpc"
	"encoding/gob"
	"log"
	"sync"
	"context"
	"errors"
	"container/list"

	"github.com/tori209/data-executor/log/format"
	"github.com/google/uuid"
)

//====================================================================

type ExecutorInfoOption struct {
	address		string
	rpcPort		string
	proto		string
	ctx			context.Context
}

// 생성은 ExecutorManager.start의 Goroutine에서 알아서 함.
type ExecutorInfo struct {
	contactURL		string
	contactProto	string
	id				uuid.UUID
	currTask		*format.TaskRequestMessage
	resChan			chan format.TaskResponseMessage
	mu				*sync.RWMutex
	ctx				context.Context
	cancel			context.CancelFunc
}

func NewExecutorInfo(opt *ExecutorInfoOption) (*ExecutorInfo) {
	if opt.ctx == nil {  opt.ctx = context.Background()  }
	ctx, cancel := context.WithCancel(opt.ctx)

	if opt.rpcPort[0] != ':' { opt.rpcPort = ":" + opt.rpcPort }

	return &ExecutorInfo{
		contactURL:		opt.address + opt.rpcPort,
		contactProto:	opt.proto,
		id:			uuid.New(),
		currTask:	nil,
		resChan:	nil, 	// requestTask 발생 시 새로 할당.
		mu:			new(sync.RWMutex),
		ctx:		ctx,
		cancel:		cancel,
	}
}

func (ei *ExecutorInfo) requestTask(request *format.TaskRequestMessage) error {
	ei.assignTask(request)

	ei.mu.RLock()

	client, err := rpc.Dial(ei.contactProto, ei.contactURL)
	if err != nil {
		ei.mu.RUnlock()
		log.Printf("[Executorinfo.requestTask] Failed to Contact Runner: %+v", err)
		ei.clearTask()
		return err
	}
	ei.mu.RUnlock()

	go func() {
		var response format.TaskResponseMessage
		err := client.Call("CodeRunner.StartTask", request, &response)
		if err != nil {
			log.Printf("[ExecutorInfo-requestTask] Connection Closed Unexpectedly: %+v", err)
			response.Status = format.TaskFailed
			response.JobID = request.JobID
			response.TaskID = request.TaskID
		}
		ei.resChan <- response
	}()
	return nil
}

func (ei *ExecutorInfo) assignTask(request *format.TaskRequestMessage) {
	ei.mu.Lock()
	defer ei.mu.Unlock()

	ei.resChan = make(chan format.TaskResponseMessage)
	ei.currTask = request
}

func (ei *ExecutorInfo) clearTask() {
	ei.mu.Lock()
	defer ei.mu.Unlock()

	close(ei.resChan)
	ei.resChan = nil
	ei.currTask = nil
}

//====================================================================

type ExecutorManager struct {
	listener	net.Listener	
	liveMap		map[string]*ExecutorInfo
	ctx			context.Context
	cancel		context.CancelFunc
	mu			*sync.RWMutex
}

func NewExecutorManager(servicePort string) (*ExecutorManager, error) {
	if servicePort[0] != ':' {
		servicePort = ":" + servicePort
	}

	listener, err := net.Listen("tcp", servicePort)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	em := ExecutorManager{
		listener:	listener,
		liveMap:	make(map[string]*ExecutorInfo),
		ctx:		ctx,
		cancel:		cancel,
		mu:			new(sync.RWMutex),
	}
	defer em.start()

	return &em, nil
}

func (em *ExecutorManager) CountOnlineExecutor() int {
	em.mu.RLock()
	ret := len(em.liveMap)
	em.mu.RUnlock()
	return ret
}


func (em *ExecutorManager) start() {
	runnerRequestProto := os.Getenv("RUNNER_REQUEST_RECEIVE_PROTO")
	if runnerRequestProto == "" {
		log.Fatalf("[Runner] RUNNER_REQUEST_RECEIVE_PROTO not defined.")
	}
	runnerRequestPort := os.Getenv("RUNNER_REQUEST_RECEIVE_PORT")
	if runnerRequestPort == "" {
		log.Fatalf("[Runner] RUNNER_REQUEST_RECEIVE_PORT not defined.")
	}

	connChan := make(chan net.Conn)
	
	// Connection Creator
	// Executor에서 요청이 들어오면 그냥 Channel을 통해 넘겨주기만.
	go func() {
		for {
			conn, err := em.listener.Accept()
			if err != nil {
				log.Println("[ExecutorManager] Failed to establish connection %+v:", err)
				continue
			}
			connChan <- conn
		}
	}()

	// Executor가 Live를 보고하면 이를 목록에 추가.
	// 실제로 Executor Info를 생성하는 영역.
	go func() {
		for {
			select{
			// Connection Clean-up
			case <-em.ctx.Done():
				em.listener.Close()
				return
				// 위의 Goroutine에서 전달한 net.Conn을 ExecutorInfo로 전환하는 부분.
			case conn := <-connChan:
				var msg format.ReportMessage
				msgDec := gob.NewDecoder(conn)
				err := msgDec.Decode(&msg);
				if err != nil {
					log.Printf("[ExecutorManager] Message From (IP: %s) Lost: %+v", err)
					conn.Close()
					continue
				}
				conn.Close()

				execInfo := NewExecutorInfo(&ExecutorInfoOption{
					address: conn.RemoteAddr().String(),
					rpcPort: runnerRequestPort,
					proto: runnerRequestProto,
					ctx: em.ctx,
				})
				
				switch msg.Kind {
				case format.RunnerStart:
					em.mu.Lock()

					em.liveMap[execInfo.contactURL] = execInfo
					log.Printf("[ExecutorManager] New Executor(IP: %s) Connected. Now %d Executors Online.",
						execInfo.contactURL,
						len(em.liveMap),
					)
	
					em.mu.Unlock()
				case format.RunnerFinish:
					em.mu.Lock()
					ei := em.liveMap[execInfo.contactURL]
					ei.clearTask()
					delete(em.liveMap, execInfo.contactURL)
					em.mu.Unlock()
				default:
					log.Printf("[ExecutorManager] Unexpected Message Received. (IP: %s / Kind: %s)",
						execInfo.contactURL,
						msg.Kind.String(),
					)
				}

				
				
			}
		}
	}()
}

func (em *ExecutorManager) ProcessJob(request format.TaskRequestMessage, sliceSize uint64) (bool, error) {
	if len(em.liveMap) == 0 {
		return false, errors.New("No available executor")
	}
	if request.RangeBegin >= request.RangeEnd {
		return false, errors.New("Invalid Argument")
	}


	// Job을 Range 단위로 쪼개어 Task로 분할
	taskList := list.New()
	for pivot := request.RangeBegin; pivot < request.RangeEnd; pivot++ {
		task := request
		task.RangeBegin = pivot
		if pivot + sliceSize <= task.RangeBegin {
			task.RangeEnd = pivot + sliceSize
		}
		taskList.PushBack(&task)
	}

	taskDoneCnt := 0
	totalTaskCnt := taskList.Len()
	diedList := make([]string, 0)
	for {
		em.mu.RLock()
		// 현재 살아있는 Executor가 존재하는가?
		if len(em.liveMap) == 0 {
			em.mu.RUnlock()
			return false, errors.New("No available executor")
		}
		// Map 순회하며 Task 할당 및 결과 확인 ----------------------------------------------------
		for id, ei := range em.liveMap {
			ei.mu.Lock()

			// 이미 작업 중, 완료 여부 확인. ======================================================
			if (ei.currTask != nil) {
				select {
				case response := <- ei.resChan: // 작업 완료
					if response.Status == format.TaskFinish {
						// 작업 성공 시 totalTaskCnt를 +1
						log.Printf("[ExecutorManager] Task Complete. (JobID: %s / TaskID: %s)", 
							response.JobID.String(),
							response.TaskID.String(),
						)
						totalTaskCnt++
					} else if response.Status == format.TaskFailed {
						// 작업 실패 시 taskList에 Request를 다시 삽입
						log.Printf("[ExecutorManager] Task Failed. Retry (JobID: %s / TaskID: %s)", 
							response.JobID.String(),
							response.TaskID.String(),
						)
						taskList.PushBack(ei.currTask)
					} else {
						// 이상한 Message 전달. 일단 통과로 치는데, 나중에 잡자.
						log.Printf("[ExecutorManager] Invalid Status Given: %+v", response)
					}
					ei.mu.Unlock()
					ei.clearTask()
					continue
					// 두 경우 모두 완료 시 Continue를 수행하여 패스시켜야 함. 
				default: // 아직 완료 X, 다음으로 이동
					ei.mu.Unlock()
					continue
				}
			}


			// 추가 할당 작업 없음. 완료 유무만 검사 ===============================================
			if (taskList.Len() == 0) {
				ei.mu.Unlock()
				continue
			} 

			// 추가 할당 작업 존재. Task 할당 시도 =================================================
			currTaskElem := taskList.Front()
			val, _ := currTaskElem.Value.(*format.TaskRequestMessage)
			if err := ei.requestTask(val); err != nil { // 네트워크 문제로 연결 실패. 
				log.Printf("[ExecutorManager] Runner Connection Lost")
				diedList = append(diedList, id)
				ei.mu.Unlock()
				continue
			} else { // 요청 성공.
				ei.mu.Unlock()
				log.Printf("[ExecutorManager] Task Assigned.")
				taskList.Remove(currTaskElem)
			}
	    }
		em.mu.RUnlock()
		// ------------------------------------------------------------------------------------------
		
		// 죽은 Connection 삭제
		if diedList != nil {
			em.mu.Lock()
			for _, id := range diedList {
				deadExec, ok := em.liveMap[id]
				if !ok {  continue  }
				deadExec.clearTask()
				delete(em.liveMap, id)
			}
			diedList = nil
			em.mu.Unlock()
		}
		// ------------------------------------------------------------------------------------------

		if (taskDoneCnt == totalTaskCnt) {  break  }
	}

	return true, nil
}

func (em *ExecutorManager) Destroy() {
	em.cancel()	// This will stop everything
}
