package fspool

import "time"

type Worker struct {
	pool *Pool
	//任务队列
	task chan func()
	//执行任务的最后时间
	lastTime time.Time
}

func (w *Worker) run() {
	go w.running()
}

func (w *Worker) running() {
	for f := range w.task {
		if f == nil {
			return
		}
		f()
		//任务完成，归还worker
		w.pool.PutWorker(w)
		w.pool.decRunning()
	}
}
