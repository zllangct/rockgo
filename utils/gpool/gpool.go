package gpool

import (
	"sync"
)

//Worker goroutine struct.
type Worker struct {
	id       int32
	p        *Pool
	jobQueue chan *Job
	stop     chan struct{}
}

//Start start gotoutine pool.
func (w *Worker) Start() {
	go func() {
		var job *Job
		for {
			select {
			case job = <-w.jobQueue:
			case job = <-w.p.jobQueue:
				//id为空时，是新连接，老链接，通过worker id 保证同一连接 由同一线程处理，减少竞态，保证顺序
				if job.WorkerID != -1 {
					if job.WorkerID >= 0 && job.WorkerID < w.p.numWorkers {
						w.p.workerQueue[job.WorkerID].jobQueue <- job
						continue
					}
				}
			case <-w.stop:
				return
			}
			//TODO 错误处理
			job.Job(job.Args...)
			w.p.jobPool.Put(job)
		}
	}()
}

//Job is a function for doing jobs.
type Job struct {
	WorkerID int32
	Args     []interface{}
	Job      func(args ...interface{})
	Callback func(id int32)
}

var globalPool *Pool

//Pool is goroutine pool config.
type Pool struct {
	numWorkers  int32
	jobQueueLen int32
	jobPool     *sync.Pool
	jobQueue    chan *Job
	workerQueue []*Worker
}

func GetGloblePool(numWorkers int, jobQueueLen int) *Pool {
	if globalPool == nil {
		globalPool = NewPool(numWorkers, jobQueueLen)
	}
	return globalPool
}

//NewPool news gotouine pool
func NewPool(numWorkers int, jobQueueLen int) *Pool {
	jobQueue := make(chan *Job, jobQueueLen)
	workerQueue := make([]*Worker, numWorkers)

	pool := &Pool{
		numWorkers:  int32(numWorkers),
		jobQueueLen: int32(jobQueueLen),
		jobQueue:    jobQueue,
		workerQueue: workerQueue,
		jobPool:     &sync.Pool{New: func() interface{} { return &Job{WorkerID: int32(-1)} }},
	}
	pool.Start()
	return pool
}

func (p *Pool) AddJobParallel(handler func(...interface{}), args []interface{}, wid int32, callback func(int32)) {
	job := p.jobPool.Get().(*Job)
	job.Job = handler
	job.Args = args
	job.Callback = callback

	p.jobQueue <- job
}

func (p *Pool) AddJobSerial(handler func(...interface{}), args []interface{}, wid int32, callback func(int32)) {
	job := p.jobPool.Get().(*Job)
	job.Job = handler
	job.Args = args
	job.Callback = callback

	if wid <= -1 || wid >= p.numWorkers {
		idStr := args[2].(string)
		sum := int32(0)
		for _, c := range idStr {
			sum = sum + int32(c)
		}
		job.WorkerID = sum % p.numWorkers
		job.Callback(job.WorkerID)
	} else {
		job.WorkerID = wid
	}

	p.workerQueue[job.WorkerID].jobQueue <- job
}

//Start starts all workers
func (p *Pool) Start() {
	for i := 0; i < cap(p.workerQueue); i++ {
		worker := &Worker{
			id:       int32(i),
			p:        p,
			jobQueue: make(chan *Job, 10),
			stop:     make(chan struct{}),
		}
		p.workerQueue[i] = worker
		worker.Start()
	}
}

func (p *Pool) Size() int32 {
	return p.numWorkers
}

//Release release all workers
func (p *Pool) Release() {
	for _, worker := range p.workerQueue {
		worker.stop <- struct{}{}
	}
}
