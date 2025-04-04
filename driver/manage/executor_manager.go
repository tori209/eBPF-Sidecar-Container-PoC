package manage

import (
	"io"
	"net"
	"log"
	"sync"
	"context"
	"errors"
	"encoding/gob"
	"container/list"

	"github.com/tori209/data-executor/log/format"
	"github.com/google/uuid"
)

//====================================================================

// 생성은 ExecutorManager.start의 Goroutine에서 알아서 함.
type ExecutorInfo struct {
	id			uuid.UUID
	conn		net.Conn
	currTask	*format.TaskRequestMessage
	resChan		chan format.TaskResponseMessage
	mu			*sync.Mutex
	ctx			context.Context
	cancel		context.CancelFunc
}

func (ei *ExecutorInfo) requestTask(request format.TaskRequestMessage) error {
	reqEnc := gob.NewEncoder(ei.conn)
	resDec := gob.NewDecoder(ei.conn)
	
	if err := reqEnc.Encode(request); err != nil {
		return err
	}
	ei.currTask = &request

	// Task로부터 결과 수신 시도.
	// 현재는 Runner가 Task 성공/실패 시점에만 전송하도록 설정되어 있음.
	// 즉, 중간에 추가 메세지가 들어올 일이 없다.
	ei.resChan = make(chan format.TaskResponseMessage)
	go func() {
		var resMsg format.TaskResponseMessage
		for {
			err := resDec.Decode(&resMsg)
			if err == nil {
				// 외부에서 요청 정상 수신
				ei.resChan <- resMsg
				return
			}
			if errors.Is(err, io.EOF) {
				// 연결 종료로 메세지 수신 실패.
				log.Printf("[ExecutorInfo-requestTask] Connection Closed Unexpectedly: %+v", err)
				resMsg.Status = format.TaskFailed
				resMsg.JobID = request.JobID
				resMsg.TaskID = request.TaskID
				ei.resChan <- resMsg
				return
			}
		}
	}()
	
	return nil
}

func (ei *ExecutorInfo) clearTask() {
	ei.mu.Lock()
	close(ei.resChan)
	ei.resChan = nil
	ei.currTask = nil
	ei.mu.Unlock()
}

//====================================================================

type ExecutorManager struct {
	listener	net.Listener	
	liveMap		map[uuid.UUID]*ExecutorInfo
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
		liveMap:	make(map[uuid.UUID]*ExecutorInfo),
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
				exec_ctx, exec_cancel := context.WithCancel(em.ctx)	

				em.mu.Lock()

				new_id := uuid.New()
				em.liveMap[new_id] = &ExecutorInfo{
					id:			new_id,
					conn: 		conn,
					resChan:	nil,
					mu:		new(sync.Mutex),
					ctx: 		exec_ctx,
					cancel: 	exec_cancel,
				}
				log.Printf("[ExecutorManager] New Executor(IP: %s) Connected. Now %d Executors Online.",
					conn.RemoteAddr().String(),
					len(em.liveMap),
				)

				em.mu.Unlock()
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
	diedList := make([]uuid.UUID, 0)
	for {
		em.mu.RLock()
		// 현재 살아있는 Executor가 존재하는가?
		if len(em.liveMap) == 0 {
			em.mu.RUnlock()
			return false, errors.New("No available executor")
		}
		// Map 순회 시도
		for id, ei := range em.liveMap {
			ei.mu.Lock()
			if (ei.currTask != nil) {
				// 이미 작업 중, 완료 여부 확인.
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
			if (taskList.Len() == 0) {
				ei.mu.Unlock()
				continue
			} // 현재 부여할 작업 X, 다음 리스트 확인

			// 할당된 작업 X + 대기 작업이 남아있음
			currTaskElem := taskList.Front()
			val, _ := currTaskElem.Value.(*format.TaskRequestMessage)
			if err := ei.requestTask(*val); err != nil { // 네트워크 문제로 연결 실패. 
				log.Printf("[ExecutorManager] Runner Connection Lost")
				diedList = append(diedList, id)
				ei.mu.Unlock()
				continue
			} else { // 요청 성공.
				ei.mu.Unlock()
				taskList.Remove(currTaskElem)
			}
	    }
		em.mu.RUnlock()
		
		// 죽은 Connection 삭제
		if diedList != nil {
			em.mu.Lock()
			for _, id := range diedList {
				deadExec, ok := em.liveMap[id]
				if !ok {  continue  }
				deadExec.conn.Close()
				deadExec.cancel()
				delete(em.liveMap, id)
			}
			diedList = nil
			em.mu.Unlock()
		}
		if (taskDoneCnt == totalTaskCnt) {  break  }
	}

	return true, nil
}

func (em *ExecutorManager) Destroy() {
	em.cancel()	// This will stop everything
}
