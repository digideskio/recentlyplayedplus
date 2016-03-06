package request

import (
	"fmt"
	"sync"
	"time"

	"github.com/hishboy/gocommons/lang"
)

// RateInfo describes a rate to be added to the limiter
type RateInfo struct {
	// The max number of requests allowed for a given time period
	max uint32
	// The server region for which this rate applies
	region string
	// Number of seconds within which the max requests can occur
	period uint32
}

// Limiter restricts the rate at which incoming LimitedDoer objects can perform
// their tasks. The rates to which limits are held can be split into distinct
// regions (as opposed to holding many limiters).
type Limiter struct {
	regions      map[string]*region
	clock        uint32
	secondTicker *time.Ticker
	lock         sync.Mutex
	isStopped    bool
}

type region struct {
	rates []*rate
	//outstanding requests for this region
	//TODO: This queue is synchronized. It doesn't need to be because sync
	// is handled at the limiter level (because of the clock). So, I'll switch
	// this out when I have time to implement my own non-thread-safe queue.
	tasks         *lang.Queue
	hasZeroPeriod bool
}

type rate struct {
	// An array of seconds, with values representing the number of requests made
	// in that second
	history []uint32
	// The number of requests that occurred in this second
	thisTick uint32
	// The remaining number of requests that can occur given time period.
	allowance uint32
	// Number of seconds within which the max requests can occur
	period uint32
}

// NewLimiter creates a Limiter object with its ticker running, and ultimately
// ready to be used with the remainder of the Limiter member functions.
func NewLimiter() *Limiter {
	ret := &Limiter{
		regions:      make(map[string]*region),
		secondTicker: time.NewTicker(1 * time.Second),
		isStopped:    false,
	}
	go ret.asyncUpdate()
	return ret
}

// Stop signals the halting of the ticker of the Limiter object and
// causes the Limiter to err on enqueues made after the ticker is successfully
// stopped . This function should be called to allow the Limiter object to be
// garbage collected. No guarantees are made that no enqueues will be admitted
// or processed between the calling of this function and the stopping of the
// ticker, but currently blocked requests will not be performed.
func (l *Limiter) Stop() {
	l.lock.Lock()
	l.isStopped = true
	l.lock.Unlock()
}

// AddRegion registers with this limiter object a new region with the called
// ${name}. Errs if the region to add is already registered with this limiter,
// or if the limiter has been stopped.
func (l *Limiter) AddRegion(name string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.isStopped {
		return fmt.Errorf("Limiter has been stopped")
	}
	_, alreadyExists := l.regions[name]
	if alreadyExists {
		return fmt.Errorf("Region %s already exists!", name)
	}
	l.regions[name] = &region{
		tasks:         lang.NewQueue(),
		rates:         nil,
		hasZeroPeriod: false,
	}
	return nil
}

// AddRate registers a new rate with the region specified within this limiter.
// A limiter will only allow a task to be performed once there is remaining
// allowance within EVERY rate for the region which the task is enqueued for.
// This errs if the region specified doesn't exist in the limiter, or if the
// limiter has been stopped. Adding a rate with a period of zero will create
// a rate which never has its allowance replenish - calls to Enqueue after this
// point will return an error.
func (l *Limiter) AddRate(limit, period uint32, region string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.isStopped {
		return fmt.Errorf("Limiter has been stopped")
	}
	reg, ok := l.regions[region]
	if !ok {
		return fmt.Errorf("Cannot add rate for unknown region '%s'", region)
	}
	reg.rates = append(reg.rates, &rate{
		history:   make([]uint32, period, period),
		thisTick:  0,
		allowance: limit,
		period:    period,
	})
	if period == 0 {
		reg.hasZeroPeriod = true
	}
	return nil
}

// Enqueue registers a LimitedDoer with the Limiter to be executed at a later
// time, when the allowance to perform the task is available. Returns a uint32
// representing the remaining allowance at the time of this function call,
// before subtracting the cost of this task. i.e. 1 means the next enqueue may
// require that the task be queued for later execution, and 0 means that the
// current task has been queued for later execution.
// Errs if the given region doesn't exist or if the Limiter has been stopped.
func (l *Limiter) Enqueue(task LimitedDoer, region string) (uint32, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.isStopped {
		return 0, fmt.Errorf("Limiter has been stopped")
	}
	var position uint32
	if reg, ok := l.regions[region]; ok && reg.allowance() > 0 {
		position = reg.allowance()
		l.regions[region].reserve()
		go l.execute(task, region)
	} else {
		if !ok {
			return 0, fmt.Errorf("Cannot queue for unknown region '%s'", region)
		} else if reg.hasZeroPeriod {
			return 0, fmt.Errorf("No more requests are allowed for region '%s'", region)
		}
		l.regions[region].tasks.Push(task)
		position = 0
	}
	return position, nil
}

// Stopped returns true if this Limiter has had Stop() called on it.
// Returns false otherwise.
func (l *Limiter) Stopped() bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.isStopped
}

func (l *Limiter) asyncUpdate() {
	//Every second, refresh all the rate objects.
	for range l.secondTicker.C {
		l.lock.Lock()
		if l.isStopped {
			l.secondTicker.Stop()
			l.lock.Unlock()
			return
		}
		for _, region := range l.regions {
			for _, rate := range region.rates {
				rate.tick(l.clock)
			}
		}
		l.clock++
		l.useAllowance()
		l.lock.Unlock()
	}
}

func (r *rate) tick(clock uint32) {
	if r.period == 0 {
		return
	}
	idx := clock % r.period
	r.allowance += r.history[idx]
	r.history[idx] = r.thisTick
	r.thisTick = 0
}

func (r *region) allowance() (allowance uint32) {
	allowance = 4294967295
	for _, rate := range r.rates {
		if rate.allowance < allowance {
			allowance = rate.allowance
		}
	}
	return allowance
}

// Decrements the allowance of each rate contained within, but does not increment
// the value of this tick. Effectively reserves an API call to this region that
// cannot be used by other calls to enqueue.
func (r *region) reserve() {
	for _, rate := range r.rates {
		rate.allowance--
	}
}

func (l *Limiter) execute(task LimitedDoer, region string) {
	task.Do()
	l.lock.Lock()
	defer l.lock.Unlock()
	r := l.regions[region]
	for _, rate := range r.rates {
		rate.thisTick++
	}
}

func (l *Limiter) useAllowance() {
	for name, r := range l.regions {
		for r.allowance() > 0 && r.tasks.Peek() != nil {
			r.reserve()
			go l.execute(r.tasks.Poll().(LimitedDoer), name)
		}
	}
}
