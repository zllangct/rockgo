package network

import (
	"math/rand"
	"sync"
)

const WORKER_ID_RANDOM int32 = -1

//Worker goroutine struct.
type Worker struct {
	id       int32
	p        *Pool
	jobQueue chan *Job
	stop     chan struct{}
}

//Start start goroutine pool.
func (w *Worker) Start() {
	go func() {
		var job *Job
		for {
			select {
			case job = <-w.jobQueue:
			case job = <-w.p.jobQueue:
				//task which worker id not nil will push into the target goroutine to insure data safety
				if job.WorkerID != WORKER_ID_RANDOM {
					if job.WorkerID >= 0 && job.WorkerID < w.p.numWorkers {
						w.p.workerQueue[job.WorkerID].jobQueue <- job
						continue
					}
				}
			case <-w.stop:
				return
			}
			//TODO handle Error
			job.Job([]interface{}{job.WorkerID},job.Args...)
			w.p.jobPool.Put(job.Init())
		}
	}()
}

//job type: Parallel;Serial
type JobType int
const JOB_TYPE_PARALLEL JobType = 1
const JOB_TYPE_SERIAL JobType = 2
const JOB_TYPE_DEFAULT = JOB_TYPE_PARALLEL

//Job is a function for doing jobs.
type Job struct {
	WorkerID int32
	Args     []interface{}
	Job      func(ctx []interface{},args ...interface{})
}

//initialize the job
func (p *Job)Init() *Job{
	p.WorkerID = -1
	p.Args = nil
	p.Job = nil
	return p
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

//get the singleton pool
func GetGlobalPool(numWorkers int, jobQueueLen int) *Pool {
	if globalPool == nil {
		globalPool = NewPool(numWorkers, jobQueueLen)
	}
	return globalPool
}

//NewPool news goroutine pool
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

//random worker, task will run in a random worker
func (p *Pool) AddJob(handler func([]interface{},...interface{}), args []interface{},typ ... JobType){
	job := p.jobPool.Get().(*Job)
	job.Job = handler
	job.Args = args
	job.WorkerID = WORKER_ID_RANDOM

	if len(typ)>0 && (typ[0] == JOB_TYPE_SERIAL){
		job.WorkerID = rand.Int31() % p.numWorkers
		p.workerQueue[job.WorkerID].jobQueue <- job
	}else{
		p.jobQueue <- job
	}
}

//fixed worker,task with the same worker id will push into the same goroutine
func (p *Pool) AddJobFixed(handler func([]interface{}, ...interface{}), args []interface{}, wid int32) {
	job := p.jobPool.Get().(*Job)
	job.Job = handler
	job.Args = args

	if wid <= -1 || wid >= p.numWorkers {
		job.WorkerID = rand.Int31() % p.numWorkers
		p.workerQueue[job.WorkerID].jobQueue <- job
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

//get the pool size
func (p *Pool) Size() int32 {
	return p.numWorkers
}

//Release release all workers
func (p *Pool) Release() {
	for _, worker := range p.workerQueue {
		worker.stop <- struct{}{}
	}
}
