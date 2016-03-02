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
	max uint64
	// The server region for which this rate applies
	region string
	// Number of seconds within which the max requests can occur
	period uint32
}

// Singleton rate limiter construct.
type limiter struct {
	regions      map[string]*region
	clock        uint64
	secondTicker <-chan time.Time
	lock         sync.Mutex
}

type region struct {
	rates []rate
	//outstanding requests for this region
	requests *lang.Queue
}

type rate struct {
	// An array of seconds, with values representing the number of requests made
	// in that second
	history []uint32
	// The max number of requests allowed for a given time period
	max uint64
	// The number of requests that occurred in this second
	thisTick uint32
	// The remaining number of requests that can occur given time period.
	allowance uint32
	// Number of seconds within which the max requests can occur
	period uint32
}

var lim *limiter

func init() {
	lim = &limiter{
		regions:      make(map[string]*region),
		secondTicker: time.Tick(1 * time.Second),
	}
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

// Requests call this. This is really the only function that should be called
// from outside of this file.
func (l *limiter) enqueue(request request) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	if reg, ok := l.regions[request.region]; ok && reg.allowance() > 0 {
		l.regions[request.region].reserve()
		go l.regions[request.region].execute(request)

	} else {
		if !ok {
			return fmt.Errorf("Attempt to queue unregistered region '%s'", request.region)
		}
		l.regions[request.region].requests.Push(request)
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

func (r *region) execute(request request) {
	request.do()
	lim.lock.Lock()
	for _, rate := range r.rates {
		rate.thisTick++
	}
	lim.lock.Unlock()
}

func (r *region) useAllowance() {
	for r.allowance() > 0 && r.requests.Peek() != nil {
		r.reserve()
		go r.execute(r.requests.Poll().(request))
	}
}
