package request_test

import (
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
				input := 1
				doer := newTestDoer(input)
				allowance, err := lim.Enqueue(doer, reg)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err here.")
				Ω(allowance).Should(Equal(uint32(0)), "Should need to queue this for later")
				output, retrieved := doer.popChannel(1)
				Ω(retrieved).Should(BeFalse(), "Should need to wait for allowance")
				output, retrieved = doer.popChannel(4)
				Ω(retrieved).Should(BeTrue(), "Should finally have completed")
				Ω(output).Should(Equal(input), "The task should have returned the input value")
			})
		})

		Context("with a rate containing a period of zero", func() {
			BeforeEach(func() {
				limit = 5
				period = 0
				lim.AddRate(limit, period, reg)
			})

			It("should allow requests until the allowance is depleted", func() {
				for i := uint32(0); i < limit; i++ {
					doer := newTestDoer(int(i))
					allowance, err := lim.Enqueue(doer, reg)
					Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err here.")
					Ω(allowance).Should(Equal(limit - i))
					output, retrieved := doer.popChannel(1)
					Ω(retrieved).Should(BeTrue(), "Should have retrieved value")
					Ω(output).Should(Equal(int(i)), "The task should have returned the input value")
				}
				input := 2
				doer := newTestDoer(input)
				allowance, err := lim.Enqueue(doer, reg)
				Ω(err).Should(HaveOccurred(), "Enqueue should err with no more possible allowance.")
				Ω(allowance).Should(Equal(uint32(0)), "Allowance should report a value of zero.")
			})
		})

		Context("with two rates assigned", func() {
			var limit1, period1, limit2, period2 uint32
			BeforeEach(func() {
				limit1, period1 = 1, 2
				limit2, period2 = 2, 6
				lim.AddRate(limit1, period1, reg)
				lim.AddRate(limit2, period2, reg)
			})

			//This is a longer test... both in runtime and code.
			It("should obey both rates", func() {
				//The first should run immediately.
				input := 0
				doer := newTestDoer(input)
				allowance, err := lim.Enqueue(doer, reg)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err on first insert.")
				Ω(allowance).Should(Equal(limit1), "Enqueue should return full allowance on first insert")
				output, retrieved := doer.popChannel(1)
				Ω(retrieved).Should(BeTrue(), "Should have retrieved value in under a second")
				Ω(output).Should(Equal(input), "The task should have returned the input value")
				//The second one should need to wait at least one second by limitation of the first rate.
				input = 1
				doer = newTestDoer(input)
				allowance, err = lim.Enqueue(doer, reg)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err on second insert.")
				Ω(allowance).Should(Equal(uint32(0)), "Enqueue should need to stall the second task")
				_, retrieved = doer.popChannel(1)
				Ω(retrieved).Should(BeFalse(), "Second task should not complete within a second.")
				// Once the first rate has allowance again, this should pass through
				output, retrieved = doer.popChannel(3)
				Ω(retrieved).Should(BeTrue(), "Second task should have completed after a delay.")
				Ω(output).Should(Equal(input), "The task should have returned the second input value after a delay")
				//Third task should need to wait for the second rate, even if the first rate has allowance
				input = 2
				doer = newTestDoer(input)
				allowance, err = lim.Enqueue(doer, reg)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err on third insert.")
				Ω(allowance).Should(Equal(uint32(0)), "Enqueue should need to stall the third task")
				_, retrieved = doer.popChannel(3)
				Ω(retrieved).Should(BeFalse(), "Third task should not complete within three seconds.")
				// But after that has passed, there should be allowance in both rates again
				output, retrieved = doer.popChannel(2)
				Ω(retrieved).Should(BeTrue(), "Third task should have completed after a longer delay.")
				Ω(output).Should(Equal(input), "The task should have returned the third input value after a longer delay")
			})
		})
	})

	//TODO: Write tests for multiple regions
	Context("When there is more than one region", func() {
		var reg1, reg2 string
		BeforeEach(func() {
			reg1, reg2 = "first", "second"
			err := lim.AddRegion(reg1)
			Ω(err).ShouldNot(HaveOccurred())
			err = lim.AddRegion(reg2)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("with one rate per region", func() {
			var limit1, period1, limit2, period2 uint32
			BeforeEach(func() {
				limit1, period1 = 1, 5
				limit2, period2 = 5, 1
				lim.AddRate(limit1, period1, reg1)
				lim.AddRate(limit2, period2, reg2)
			})

			It("should only obey the first rate for the first region", func() {
				input := 0
				doer := newTestDoer(input)
				allowance, err := lim.Enqueue(doer, reg1)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err on first insert.")
				Ω(allowance).Should(Equal(limit1), "Enqueue should return full allowance on first insert")
				output, retrieved := doer.popChannel(1)
				Ω(retrieved).Should(BeTrue(), "Should have retrieved value in under a second")
				Ω(output).Should(Equal(input), "The task should have returned the input value")
				input = 1
				doer = newTestDoer(input)
				allowance, err = lim.Enqueue(doer, reg1)
				Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err on second insert.")
				Ω(allowance).Should(Equal(uint32(0)), "Enqueue should need to stall the second task")
				_, retrieved = doer.popChannel(2)
				Ω(retrieved).Should(BeFalse(), "Second task should not complete within a second.")
			})

			It("should only obey the second rate for the second region", func() {
				for i := uint32(0); i < limit2; i++ {
					doer := newTestDoer(int(i))
					allowance, err := lim.Enqueue(doer, reg2)
					Ω(err).ShouldNot(HaveOccurred(), "Enqueue shouldn't err on insert.")
					Ω(allowance).Should(Equal(limit2-i), "Enqueue should return correct allowance on insert")
					output, retrieved := doer.popChannel(1)
					Ω(retrieved).Should(BeTrue(), "Should have retrieved value in under a second")
					Ω(output).Should(Equal(int(i)), "The task should have returned the input value")
				}
			})
		})

	})

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
