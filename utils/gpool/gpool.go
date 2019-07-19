package gpool

import (
	"sync"
)

//Worker goroutine struct.
type Worker struct {
	id   int32
	p    *Pool
	jobQueue chan *Job
	stop chan struct{}
}

//Start start gotoutine pool.
func (w *Worker) Start() {
	length:=w.p.Size()
	go func() {
		var job *Job
		for {
			select {
			case job=<-w.jobQueue:
			case job = <-w.p.JobQueue:
				//id为空时，是新连接，老链接，通过worker id 保证同一连接 由同一线程处理，减少竞态，保证顺序
				if job.WorkerID != -1 && job.WorkerID < int32(length) {
					w.p.WorkerQueue[job.WorkerID].jobQueue <- job
					continue
				}
			case <-w.stop:
				return
			}
			//TODO 错误处理
			job.Job(job.Args...)
			job.Callback(w.id)
			job.WorkerID = int32(-1)
			w.p.JobPool.Put(job)
		}
	}()
}

//Job is a function for doing jobs.
type Job struct {
	WorkerID int32
	Args []interface{}
	Job      func(args ...interface{})
	Callback func(id int32)
}

var globalPool *Pool
//Pool is goroutine pool config.
type Pool struct {
	JobPool     *sync.Pool
	JobQueue    chan *Job
	WorkerQueue []*Worker
}

func GetGloblePool(numWorkers int, jobQueueLen int) *Pool {
	if globalPool == nil {
		globalPool = NewPool(numWorkers,jobQueueLen)
	}
	return globalPool
}

//NewPool news gotouine pool
func NewPool(numWorkers int, jobQueueLen int) *Pool {
	jobQueue := make(chan *Job, jobQueueLen)
	workerQueue := make([]*Worker, numWorkers)

	pool := &Pool{
		JobQueue:    jobQueue,
		WorkerQueue: workerQueue,
		JobPool:     &sync.Pool{New: func() interface{} { return &Job{WorkerID:int32(-1)} }},
	}
	pool.Start()
	return pool
}

func (p *Pool)AddJob(handler func(...interface{}), args []interface{},wid int32,callback func(int32))  {
	job:=p.JobPool.Get().(*Job)
	job.Job = handler
	job.Args = args
	job.Callback = callback
	job.WorkerID =wid
	p.JobQueue <-job
}

//Start starts all workers
func (p *Pool) Start() {
	for i := 0; i < cap(p.WorkerQueue); i++ {
		worker := &Worker{
			id:   int32(i),
			p:    p,
			jobQueue:make(chan *Job,10),
			stop: make(chan struct{}),
		}
		p.WorkerQueue[i] = worker
		worker.Start()
	}
}

func (p *Pool)Size() int {
	return len(p.WorkerQueue)
}

//Release release all workers
func (p *Pool) Release() {
	for _, worker := range p.WorkerQueue {
		worker.stop<- struct{}{}
	}
}
