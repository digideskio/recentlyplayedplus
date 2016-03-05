package request_test

import (
	"fmt"
	"time"

	. "github.com/thomasmmitchell/recentlyplayedplus/request"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type TestDoer struct {
	channel chan int
	value   int
}

func newTestDoer(value int) *TestDoer {
	return &TestDoer{
		channel: make(chan int, 1),
		value:   value,
	}
}

func (td *TestDoer) Do() {
	td.channel <- td.value
}

//int is the value from the channel. bool is whether or not the value
// was retrieved before timing out.
func (td *TestDoer) popChannel(timeout time.Duration) (int, bool) {
	select {
	case v := <-(td.channel):
		return v, true
	case <-time.After(timeout * time.Second):
		return 0, false
	}
}

//=======TESTS START HERE=========//

var _ = Describe("Limiter", func() {
	var lim *Limiter

	BeforeEach(func() {
		lim = NewLimiter()
		Ω(lim).ShouldNot(BeNil(), "A new limiter should have been allocated.")
	})

	AfterEach(func() {
		lim.Stop()
	})

	Context("When there are no regions", func() {
		It("should err for any enqueue", func() {
			doer0 := newTestDoer(0)
			_, err := lim.Enqueue(doer0, "NA")
			Ω(err).Should(HaveOccurred(), "Enqueuing with no regions should error")
		})

	})

	Context("When there is a single region", func() {
		reg := "test1"
		notreg := "test2"

		BeforeEach(func() {
			err := lim.AddRegion(reg)
			Ω(err).ShouldNot(HaveOccurred(), "Should be able to add a region")
		})

		It("should err if the same region is added again", func() {
			err := lim.AddRegion(reg)
			Ω(err).Should(HaveOccurred(), "Shouldn't be able to add the same region twice")
		})

		Context("with no rates set", func() {
			It("should complete a single task", func() {
				input := 0
				doer := newTestDoer(input)
				allowance, err := lim.Enqueue(doer, reg)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err here")
				Ω(allowance > 0).Should(BeTrue(), "Allowance should not imply that future tasks are blocked.")
				output, retrieved := doer.popChannel(2)
				Ω(retrieved).Should(BeTrue(), "Should have retrieved value ")
				Ω(output).Should(Equal(input), "The task should have returned")
			})

		})

		It("should err if enqueue is attempted for a non-existant region", func() {
			doer := newTestDoer(0)
			_, err := lim.Enqueue(doer, notreg)
			Ω(err).Should(HaveOccurred(), "Should not be allowed to enqueue for non-existant region")
		})

		var limit, period uint32

		testSingleTask := func() {
			input := 0
			doer := newTestDoer(input)
			allowance, err := lim.Enqueue(doer, reg)
			Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err here.")
			Ω(allowance).Should(Equal(limit))
			output, retrieved := doer.popChannel(1)
			Ω(retrieved).Should(BeTrue(), "Should have retrieved value")
			Ω(output).Should(Equal(input), "The task should have returned the input value")
		}

		Context("with a single large rate", func() {
			BeforeEach(func() {
				limit = 10000
				period = 20
				lim.AddRate(limit, period, reg)
			})

			It("should complete a single task", func() {
				testSingleTask()
			})

			It("should complete a whole bunch of tasks", func() {
				numTasks := 25
				doers := make([]*TestDoer, 0, numTasks)
				for i := 0; i < numTasks; i++ {
					doers = append(doers, newTestDoer(i))
					lim.Enqueue(doers[i], reg)
				}
				for i := 0; i < numTasks; i++ {
					value, retrieved := doers[i].popChannel(1)
					Ω(retrieved).Should(BeTrue())
					Ω(value).Should(Equal(i))
				}
			})

		})
		Context("with a single small rate", func() {
			BeforeEach(func() {

				limit = 1
				period = 3
				lim.AddRate(limit, period, reg)
			})

			It("should complete a single task", func() {
				testSingleTask()
			})

			It("should stall for a task after the allowance is depleted", func() {
				testSingleTask()
				fmt.Println("check0")
				input := 1
				doer := newTestDoer(input)
				allowance, err := lim.Enqueue(doer, reg)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err here.")
				Ω(allowance).Should(Equal(uint32(0)), "Should need to queue this for later")
				output, retrieved := doer.popChannel(1)
				Ω(retrieved).Should(BeFalse(), "Should need to wait for allowance")
				output, retrieved = doer.popChannel(3)
				Ω(retrieved).Should(BeTrue(), "Should finally have completed")
				Ω(output).Should(Equal(input), "The task should have returned the input value")
			})
		})
	})

	//TODO: Write tests for multiple regions

	Context("When the limiter has been stopped", func() {
		reg := "NA"
		BeforeEach(func() {
			err := lim.AddRegion(reg)
			Ω(err).ShouldNot(HaveOccurred(), "Should be able to add region before stop.")
			lim.Stop()
		})

		It("should not allow further enqueueing", func() {
			doer := newTestDoer(0)
			_, err := lim.Enqueue(doer, reg)
			Ω(err).Should(HaveOccurred(), "Should not be able to enqueue after stop")
			_, retrieved := doer.popChannel(1)
			Ω(retrieved).Should(BeFalse(), "The task should not be completed.")
		})

		It("should not allow more regions to be added", func() {
			err := lim.AddRegion("We are region")
			Ω(err).Should(HaveOccurred(), "Should not be able to add regions after stop")
		})

		It("should not allow more rates to be added", func() {
			err := lim.AddRate(100, 100, reg)
			Ω(err).Should(HaveOccurred(), "Should not be able to add rates after stop")
		})
	})
})
