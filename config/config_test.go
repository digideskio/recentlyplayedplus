package config_test

import (
	"os"
	"reflect"
	"sort"

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

func numRatesIs(expected int) bool {
	return len(Rates()) == expected
}

func regionsAreCorrect(expected []string) bool {
	if !numRegionsIs(len(expected)) {
		return false
	}
	configured := Regions()
	sort.Strings(configured)
	sort.Strings(expected)
	for i, v := range configured {
		if v != expected[i] {
			return false
		}
	}
	return true
}

func ratesAreCorrect(expected []types.Rate) bool {
	if !numRatesIs(len(expected)) {
		return false
	}
	configured := Rates()
	// O(n^2), because I don't want to implement sort for types.Rate.
	for _, r1 := range configured {
		for _, r2 := range expected {
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
		var expectedRegions []string

		Context("With one region", func() {
			BeforeEach(func() {
				configFile = "oneregonerate.yml"
				expectedRegions = []string{"na"}
			})

			It("should register only one region", func() {
				Ω(numRegionsIs(1)).Should(BeTrue())
			})

			Specify("the region should have the expected value", func() {
				Ω(regionsAreCorrect(expectedRegions)).Should(BeTrue())
			})
		})

		Context("With many regions", func() {
			BeforeEach(func() {
				configFile = "manyregsonerate.yml"
				expectedRegions = []string{"na", "euw", "kr", "eune", "lan", "las", "oce", "tr", "ru", "pbe"}
			})

			It("should register ten regions", func() {
				Ω(numRegionsIs(10)).Should(BeTrue())
			})

			Specify("each rate should have the correct constraints", func() {
				Ω(regionsAreCorrect(expectedRegions)).Should(BeTrue())
			})
		})
	})

	Context("When loading rates", func() {
		var expectedRates []types.Rate

		Context("With one rate", func() {
			BeforeEach(func() {
				configFile = "oneregonerate.yml"

				expectedRates = []types.Rate{{Period: 10, Max: 10}}
			})

			It("should have one rate", func() {
				Ω(numRatesIs(1)).Should(BeTrue())
			})

			Specify("the rate should have the correct constraints", func() {
				Ω(ratesAreCorrect(expectedRates)).Should(BeTrue())
			})
		})
		Context("With many rates", func() {
			BeforeEach(func() {
				configFile = "oneregmanyrates.yml"
				expectedRates = []types.Rate{
					{Period: 10, Max: 10},
					{Period: 600, Max: 500},
					{Period: 3, Max: 2},
				}
			})

			It("should have 3 rates", func() {
				Ω(numRatesIs(3)).Should(BeTrue())
			})

			Specify("the rate should have the correct constraints", func() {
				Ω(ratesAreCorrect(expectedRates)).Should(BeTrue())
			})
		})
	})
})
