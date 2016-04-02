package config_test

import (
	"os"
	"reflect"

	. "github.com/thomasmmitchell/recentlyplayedplus/config"
	"github.com/thomasmmitchell/recentlyplayedplus/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var configPrefix string = os.Getenv("GOPATH") + "/src/github.com/thomasmmitchell/recentlyplayedplus/raw/testconfs/output/"
var configFile string

func numRegionsIs(expected int) bool {
	return len(Regions()) == expected
}

//Only supports all regions having the same number of rates, which is fine for tests.
func numRatesIs(expected int) bool {
	for _, rates := range Regions() {
		if len(rates) != expected {
			return false
		}
	}
	return true
}

func registeredRegionsMatch(expected map[string][]types.Rate) bool {
	if !numRegionsIs(len(expected)) {
		return false
	}
	configured := Regions()
	for reg, rates := range configured {
		exprates, ok := expected[reg]
		if !ok {
			return false
		}
		if !containsSameRates(rates, exprates) {
			return false
		}
	}
	return true
}

func containsSameRates(first, second []types.Rate) bool {
	if len(first) != len(second) {
		return false
	}
	for _, r1 := range first {
		for _, r2 := range second {
			if reflect.DeepEqual(r1, r2) {
				goto next
			}
		}
		//If we get here, there's a missing entry
		return false
	next: //Skip out of that nested loop
	}
	return true
}

var _ = Describe("Config", func() {

	JustBeforeEach(func() {
		err := LoadConfig(configPrefix + configFile)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("When loading regions", func() {
		var expectedRegions map[string][]types.Rate

		Context("With one region", func() {

			Context("With one rate", func() {

				BeforeEach(func() {
					configFile = "onereg.yml"

					expectedRegions = make(map[string][]types.Rate)
					expectedRegions["na"] = []types.Rate{
						{Period: 10, Max: 10},
					}
				})

				It("should register only one region", func() {
					Ω(numRegionsIs(1)).Should(BeTrue())
				})

				It("should only have one rate", func() {
					Ω(numRatesIs(1)).Should(BeTrue())
				})

				Specify("the rate should have the correct constraints", func() {
					Ω(registeredRegionsMatch(expectedRegions)).Should(BeTrue())
				})
			})

			Context("With many rates", func() {
				BeforeEach(func() {
					configFile = "manyrates.yml"

					expectedRegions = make(map[string][]types.Rate)
					expectedRegions["na"] = []types.Rate{
						{Period: 10, Max: 10},
						{Period: 600, Max: 500},
						{Period: 3, Max: 2},
					}
				})

				It("should register only one region", func() {
					Ω(numRegionsIs(1)).Should(BeTrue())
				})

				It("should have registered three rates", func() {
					Ω(numRatesIs(3)).Should(BeTrue())
				})

				Specify("all three rates should have the correct constraints", func() {
					Ω(registeredRegionsMatch(expectedRegions)).Should(BeTrue())
				})
			})
		})

		Context("With many regions", func() {
			Context("With one rate each", func() {
				BeforeEach(func() {
					configFile = "manyregs.yml"

					expectedRegions = make(map[string][]types.Rate)
					expectedRegions["na"] = []types.Rate{
						{Period: 10, Max: 10},
					}
					expectedRegions["euw"] = []types.Rate{
						{Period: 11, Max: 9},
					}
					expectedRegions["kr"] = []types.Rate{
						{Period: 12, Max: 8},
					}
				})

				It("should register three regions", func() {
					Ω(numRegionsIs(3)).Should(BeTrue())
				})

				Specify("each region should have only one rate", func() {
					Ω(numRatesIs(1)).Should(BeTrue())
				})

				Specify("each rate should have the correct constraints", func() {
					Ω(registeredRegionsMatch(expectedRegions)).Should(BeTrue())
				})
			})

			Context("With many rates each", func() {
				BeforeEach(func() {
					configFile = "manyregsmanyrates.yml"

					expectedRegions = make(map[string][]types.Rate)
					expectedRegions["na"] = []types.Rate{
						{Period: 10, Max: 10},
						{Period: 600, Max: 500},
						{Period: 3, Max: 2},
					}
					expectedRegions["euw"] = []types.Rate{
						{Period: 11, Max: 9},
						{Period: 601, Max: 499},
						{Period: 4, Max: 1},
					}
					expectedRegions["kr"] = []types.Rate{
						{Period: 12, Max: 8},
						{Period: 602, Max: 498},
						{Period: 5, Max: 0},
					}
				})

				It("should register three regions", func() {
					Ω(numRegionsIs(3)).Should(BeTrue())
				})

				Specify("each region should have three rates", func() {
					Ω(numRatesIs(3)).Should(BeTrue())
				})

				Specify("each rate should have the correct constraints", func() {
					Ω(registeredRegionsMatch(expectedRegions)).Should(BeTrue())
				})
			})
		})
	})
})
