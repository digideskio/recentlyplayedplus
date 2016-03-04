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
	secondTicker <-chan time.Time
	lock         sync.Mutex
}

type region struct {
	rates []rate
	//outstanding requests for this region
	//TODO: This queue is synchronized. It doesn't need to be because sync
	// is handled at the limiter level (because of the clock). So, I'll switch
	// this out when I have time to implement my own non-thread-safe queue.
	tasks *lang.Queue
}

type rate struct {
	// An array of seconds, with values representing the number of requests made
	// in that second
	history []uint32
	// The max number of requests allowed for a given time period
	max uint32
	// The number of requests that occurred in this second
	thisTick uint32
	// The remaining number of requests that can occur given time period.
	allowance uint32
	// Number of seconds within which the max requests can occur
	period uint32
}

var lim *Limiter

func init() {
	lim = NewLimiter()
	go func() {
		//Every second, refresh all the rate objects.
		for range lim.secondTicker {
			lim.lock.Lock()
			defer lim.lock.Unlock()
			for _, region := range lim.regions {
				for _, rate := range region.rates {
					rate.tick()
				}
				region.useAllowance()
			}
			lim.clock++
		}
	}()
}

func (r *rate) tick() {
	idx := lim.clock % r.max
	r.allowance += r.history[idx]
	r.history[idx] = r.thisTick
	r.thisTick = 0
}

func NewLimiter() *Limiter {
	return &Limiter{
		regions:      make(map[string]*region),
		secondTicker: time.Tick(1 * time.Second),
	}
}

// Add Region registers with this limiter object a new region with the called
// ${name}. Errs if the region to add is already registered with this limiter.
func (l *Limiter) AddRegion(name string) error {
	_, alreadyExists := l.regions[name]
	if alreadyExists {
		return fmt.Errorf("Region %s already exists!", name)
	}
	l.regions[name] = &region{
		tasks: lang.NewQueue(),
		rates: nil,
	}
	return nil
}

func (l *Limiter) AddRate(limit, period uint32, region string) error {
	reg, ok := l.regions[region]
	if !ok {
		return fmt.Errorf("Cannot add rate for unknown region '%s'", region)
	}
	reg.rates = append(reg.rates, rate{
		history:   make([]uint32, period, period),
		max:       limit,
		thisTick:  0,
		allowance: limit,
		period:    period,
	})
	return nil
}

// Enqueue registers a LimitedDoer with the Limiter to be executed at a later
// time, when the allowance to perform the task is available.
func (l *Limiter) Enqueue(task LimitedDoer) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if reg, ok := l.regions[task.Region()]; ok && reg.allowance() > 0 {
		l.regions[task.Region()].reserve()
		go l.regions[task.Region()].execute(task)

	} else {
		if !ok {
			return fmt.Errorf("Cannot queue for unknown region '%s'", task.Region())
		}
		l.regions[task.Region()].tasks.Push(task)
	}
	return nil
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

func (r *region) execute(task LimitedDoer) {
	task.Do()
	lim.lock.Lock()
	for _, rate := range r.rates {
		rate.thisTick++
	}
	lim.lock.Unlock()
}

func (r *region) useAllowance() {
	for r.allowance() > 0 && r.tasks.Peek() != nil {
		r.reserve()
		go r.execute(r.tasks.Poll().(LimitedDoer))
	}
}
