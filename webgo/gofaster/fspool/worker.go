package fspool

import (
	fslog "gofaster/log"
	"time"
)

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
	defer func() {
		w.pool.workerCache.Put(w)
		w.pool.decRunning()
		if err := recover(); err != nil {
			//捕获任务发生的错误
			if w.pool.panicHandler != nil {
				w.pool.panicHandler()
			} else {
				fslog.Default().Error(err)
			}
		}
	}()
	for f := range w.task {
		if f == nil {
			//回收入缓存池
			return
		}
		f()
		//任务完成，归还worker
		w.pool.PutWorker(w)
	}
}
