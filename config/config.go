package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/thomasmmitchell/recentlyplayedplus/types"
)

//Config holds information from a configuration file.
type config struct {
	//APIKey is the Riot Games API Key that tracks your apps API calls
	APIKey string
	//The applicable regions for requests to be made in
	Regions map[string][]types.Rate
}

var conf config

func LoadConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(buf, &conf)
	return err
}

func ApiKey() string {
	return conf.APIKey
}

func Regions() map[string][]types.Rate {
	return conf.Regions
}
