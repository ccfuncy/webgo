package fspool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const DefaultExpire = 1

type sig struct {
}

type Pool struct {
	//池容量
	cap int32
	//空闲worker
	workers []*Worker
	//正在运行的worker的数量
	running int32
	//过期时间 空闲的worker超过这个时间就回收掉
	expire time.Duration
	//关闭信号
	release chan sig
	//保护pool资源相关安全
	lock sync.Mutex
	//释放只能调用一次
	once sync.Once
	//workers缓存
	workerCache sync.Pool
	//cond
	cond *sync.Cond
	//错误处理
	panicHandler func()
}

func NewTimePool(cap int32, expire int) (*Pool, error) {
	if cap <= 0 {
		return nil, errors.New("pool cap can not <0")
	}
	if expire <= 0 {
		return nil, errors.New("pool expire can not <0")
	}
	p := &Pool{cap: cap,
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1)}
	p.workerCache.New = func() any {
		return &Worker{
			pool: p,
			task: make(chan func()),
		}
	}
	p.cond = sync.NewCond(&p.lock)
	go p.expireWorker()
	return p, nil
}

func NewPool(cap int32) (*Pool, error) {
	return NewTimePool(cap, DefaultExpire)
}

func (p *Pool) Submit(task func()) error {
	if len(p.release) > 0 {
		return errors.New("pool has been released!!")
	}
	//获取池子里面的worker，然后执行
	w := p.GetWorker()
	w.task <- task
	p.incRunning()
	return nil
}

func (p *Pool) GetWorker() *Worker {
	// 获取pool中的worker
	// 如果有空闲Worker 直接获取
	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n >= 0 {
		w := idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return w
	}
	// 如果没有空闲的worker,要新建一个worker
	if p.running < p.cap {
		p.lock.Unlock()
		//还不够pool的容量，直接新建一个
		c := p.workerCache.Get()
		var w *Worker
		if c == nil {
			w = &Worker{pool: p, task: make(chan func(), 1)}
		} else {
			w = c.(*Worker)
		}
		w.run()
		return w
	}
	p.lock.Unlock()
	// 如果正在运行的workers 如果大于cap 阻塞等待，worker释放
	return p.waitIdleWorker()
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) PutWorker(w *Worker) {
	//now := time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.cond.Signal()
	p.lock.Unlock()
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) Release() {
	p.once.Do(func() {
		//只执行一次
		p.lock.Lock()
		defer p.lock.Unlock()
		for i, worker := range p.workers {
			worker.task = nil
			worker.pool = nil
			p.workers[i] = nil
		}
		p.workers = nil
		p.release <- sig{}
	})
}

func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = p.release
	go p.expireWorker()
	return true
}

func (p *Pool) expireWorker() {
	//定期清除过期空闲的worker
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClosed() {
			break
		}
		p.lock.Lock()
		n := len(p.workers) - 1
		//循环空闲的worker，如果当前时间和worker最后运行任务的时间差值大于expire时，进行清理、
		for i, worker := range p.workers {
			if time.Now().Sub(worker.lastTime) <= p.expire {
				return
			} else {
				worker.task <- nil
				n = i
			}
		}
		if n >= len(p.workers)-1 {
			p.workers = p.workers[:0]
		} else {
			p.workers = p.workers[n+1:]
		}
		fmt.Printf("清除成功，running:%d\n", p.running)
		p.lock.Unlock()
	}
}

func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	p.cond.Wait()
	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		p.lock.Unlock()
		return p.waitIdleWorker()
	}
	w := idleWorkers[n]
	idleWorkers[n] = nil
	p.workers = idleWorkers[:n]
	p.lock.Unlock()
	return w
}
