package breaker

import (
	"errors"
	"gofaster/log"
	"sync"
	"time"
)

// 熔断器
type State int

const (
	StateClosed   State = iota //连续成功达到阈值
	StateHalfOpen              //连续失败超时
	StateOpen                  //连续失败达到阈值
)

// 计数
type Counts struct {
	Requests           uint32 //请求数量
	TotalSuccess       uint32 //总成功数
	TotalFailure       uint32 //总失败数
	ConsecutiveSuccess uint32 //连续成功数
	ConsecutiveFailure uint32 //连续失败数
}

func (c *Counts) OnRequest() {
	c.Requests += 1
}

func (c *Counts) OnSuccess() {
	c.TotalSuccess += 1
	c.ConsecutiveSuccess += 1
	c.ConsecutiveFailure = 0
}

func (c *Counts) OnFailure() {
	c.TotalFailure += 1
	c.ConsecutiveFailure += 1
	c.ConsecutiveSuccess = 0
}

func (c *Counts) Clear() {
	c.Requests = 0
	c.TotalFailure = 0
	c.TotalSuccess = 0
	c.ConsecutiveFailure = 0
	c.ConsecutiveSuccess = 0
}

type Setting struct {
	Name          string                            //名字
	MaxRequests   uint32                            //最大请求数，当连续成功数大于此时，断路器关闭
	Interval      time.Duration                     //间隔时间，按间隔计数，每次时间间隔统计
	Timeout       time.Duration                     //超时时间
	ReadyToTrip   func(counts Counts) bool          //是否熔断
	IsSuccessful  func(err error) bool              //是否成功
	OnStateChange func(name string, from, to State) //状态变更
	Fallback      func(error) (any, error)          //降级方案

}

type CircuitBreaker struct {
	name          string                            //名字
	maxRequests   uint32                            //最大请求数，当连续成功数大于此时，断路器关闭
	interval      time.Duration                     //间隔时间
	timeout       time.Duration                     //超时时间,
	readyToTrip   func(counts Counts) bool          //是否熔断
	isSuccessful  func(err error) bool              //是否成功
	onStateChange func(name string, from, to State) //状态变更

	mutex      sync.Mutex
	state      State
	generation uint64 //
	counts     Counts
	expire     time.Time //到期时间，开到半开
	fallback   func(error) (any, error)
}

func (b *CircuitBreaker) NewGeneration() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.counts.Clear()
	b.generation++
	var zero time.Time
	switch b.state {
	case StateClosed:
		if b.interval == 0 {
			b.expire = zero
		} else {
			b.expire = time.Now().Add(b.interval)
		}
	case StateOpen:
		b.expire = time.Now().Add(b.timeout)
	case StateHalfOpen:
		b.expire = zero
	}
}

func (b *CircuitBreaker) Execute(req func() (any, error)) (any, error) {
	//请求之前做判断，是否执行断路器
	err, generate := b.onBeforeRequest()
	if err != nil {
		//执行降级方案
		if b.fallback != nil {
			return b.fallback(err)

		}
		return nil, err
	}
	result, err := req()
	b.counts.OnRequest()
	//请求之后，是否执行断路器
	b.afterRequest(b.isSuccessful(err), generate)
	return result, err
}

func (b *CircuitBreaker) onBeforeRequest() (error, uint64) {
	//判断断路器状态，如果断路器打开，直接返回error
	state, generate := b.currentState(time.Now())
	if state == StateOpen {
		return errors.New("熔断器打开"), generate
	}
	if state == StateHalfOpen {
		if b.counts.Requests > b.maxRequests {
			return errors.New("请求过多"), generate
		}
	}
	return nil, generate
}

func (b *CircuitBreaker) afterRequest(successful bool, before uint64) {
	state, _ := b.currentState(time.Now())
	if before != b.generation {
		return
	}
	if successful {
		b.OnSuccess(state)
	} else {
		b.OnFailure(state)
	}
}

func (b *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch b.state {
	case StateClosed:
		//当前间隔完比，开始下一个间隔
		if !b.expire.IsZero() && b.expire.Before(now) {
			b.NewGeneration()
		}
	//忽略
	case StateOpen:
		//超时则半开
		if b.expire.Before(now) {
			b.SetState(StateHalfOpen)
		}
	}
	return b.state, b.generation
}

func (b *CircuitBreaker) SetState(target State) {
	if target == b.state {
		return
	}
	before := b.state
	b.state = target
	b.NewGeneration()
	if b.onStateChange != nil {
		b.onStateChange(b.name, before, target)
	}
}

func (b *CircuitBreaker) OnSuccess(state State) {
	b.counts.OnSuccess()
	switch state {
	case StateHalfOpen:
		if b.counts.ConsecutiveSuccess > b.maxRequests {
			b.SetState(StateClosed)
		}
	}
}

func (b *CircuitBreaker) OnFailure(state State) {
	b.counts.OnFailure()
	switch state {
	case StateClosed:
		if b.counts.ConsecutiveFailure > b.maxRequests {
			b.SetState(StateOpen)
		}
	case StateHalfOpen:
		if b.readyToTrip(b.counts) {
			b.SetState(StateOpen)
		}
	}
}

func NewCircuitBreaker(st Setting) *CircuitBreaker {
	c := &CircuitBreaker{
		name:          st.Name,
		onStateChange: st.OnStateChange,
	}
	if st.MaxRequests == 0 {
		c.maxRequests = 1
	} else {
		c.maxRequests = st.MaxRequests
	}

	if st.Interval == 0 {
		c.interval = time.Duration(0) * time.Second
	} else {
		c.interval = st.Interval
	}
	//断路器 开->半开
	if st.Timeout == 0 {
		c.timeout = time.Duration(20) * time.Second
	} else {
		c.timeout = st.Timeout
	}
	if st.ReadyToTrip == nil {
		c.readyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailure > 5
		}
	} else {
		c.readyToTrip = st.ReadyToTrip
	}
	if st.IsSuccessful == nil {
		c.isSuccessful = func(err error) bool {
			return err == nil
		}
	} else {
		c.isSuccessful = st.IsSuccessful
	}
	if st.Fallback == nil {
		c.fallback = func(err error) (any, error) {
			log.Default().Error("执行降级方案")
			return "降级方案", nil
		}
	} else {
		c.fallback = st.Fallback
	}

	c.NewGeneration()
	return c
}
