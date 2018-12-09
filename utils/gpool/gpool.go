package gpool

import "sync"

//Worker goroutine struct.
type Worker struct {
	id       int32
	p        *Pool
	JobQueue chan *Job
	Stop     chan struct{}
}

//Start start gotoutine pool.
func (w *Worker) Start() {
	go func() {
		var job *Job
		for {
			select {
			case job = <-w.p.JobDispatch:
			case job = <-w.JobQueue:
			case <-w.Stop:
				w.Stop <- struct{}{}
				return
			}
			job.Job()
			if job.Callback!=nil {
				job.Callback(w.id)
			}
			job.Callback=nil
			job.WorkerID = -1
			w.p.JobPool.Put(job)
		}
	}()
}

//Job is a function for doing jobs.
type Job struct {
	WorkerID int32
	Job      func()
	Callback func(w int32)
}

//Pool is goroutine pool config.
type Pool struct {
	JobPool     *sync.Pool
	JobQueue    chan *Job
	JobDispatch chan *Job
	WorkerQueue []*Worker
	stop        chan struct{}
}

//NewPool news gotouine pool
func NewPool(numWorkers int, jobQueueLen int) *Pool {
	jobQueue := make(chan *Job, jobQueueLen)
	jobDispatch := make(chan *Job, jobQueueLen)
	workerQueue := make([]*Worker, numWorkers)

	pool := &Pool{
		JobDispatch: jobDispatch,
		JobQueue:    jobQueue,
		WorkerQueue: workerQueue,
		stop:        make(chan struct{}),
		JobPool:     &sync.Pool{New: func() interface{} { return &Job{WorkerID: -1} }},
	}
	pool.Start(jobQueueLen)
	return pool
}

//Start starts all workers
func (p *Pool) Start(jobQueueLen int) {
	for i := 0; i < cap(p.WorkerQueue); i++ {
		worker := &Worker{
			id:       int32(i),
			JobQueue: make(chan *Job, jobQueueLen),
			p:        p,
			Stop:     make(chan struct{}),
		}
		worker.Start()
	}

	go p.dispatch()
}

func (p *Pool) dispatch() {
	length := len(p.WorkerQueue)
	for {
		select {
		case job := <-p.JobQueue:
			//id为空时，是新连接，老链接，通过worker id 保证同一连接 由同一线程处理，减少竞态，保证顺序
			if job.WorkerID != -1 && job.WorkerID < int32(length) {
				p.WorkerQueue[job.WorkerID].JobQueue <- job
				p.JobDispatch <- job
			}else{
				p.JobDispatch<-job
			}
		case <-p.stop:
			for _, v := range p.WorkerQueue {
				v.Stop <- struct{}{}
			}
			p.stop <- struct{}{}
			return
		}
	}
}

//Release release all workers
func (p *Pool) Release() {
	p.stop <- struct{}{}
	<-p.stop
}
