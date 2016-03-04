package request_test

import (
	. "github.com/thomasmmitchell/recentlyplayedplus/request"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestDoer struct {
	region  string
	channel chan int
}

func newTestDoer(region string) *TestDoer {
	return &TestDoer{
		region:  region,
		channel: make(chan int),
	}
}

func (td *TestDoer) Do() {
	//TODO
}

func (td *TestDoer) Region() string {
	return td.region
}

var _ = Describe("Limiter", func() {
	var lim *Limiter
	BeforeEach(func() {
		lim = NewLimiter()
	})

	Context("When there are no regions", func() {
		It("should err for any enqueue", func() {
			//TODO
		})
	})

	Context("When there is a single region", func() {
		reg := "test1"
		BeforeEach(func() {
			err := lim.AddRegion(reg)
			Ω(err).ShouldNot(HaveOccurred(), "Should be able to add a region")
		})
		It("should err if the same region is added again", func() {
			err := lim.AddRegion(reg)
			Ω(err).Should(HaveOccurred(), "Shouldn't be able to add the same region twice")
		})
		Context("with a single adequate rate", func() {
			BeforeEach(func() {
				lim.AddRate(10000, 1, reg)
			})
			It("should complete a single task", func() {
				doer0 := newTestDoer(reg)
				lim.Enqueue(doer0)
				//TODO Check response
			})
		})
	})
})
