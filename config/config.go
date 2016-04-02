package config

import "github.com/thomasmmitchell/recentlyplayedplus/types"

//Config holds information from a configuration file.
type Config struct {
	//APIKey is the Riot Games API Key that tracks your apps API calls
	APIKey string
	//The applicable regions for requests to be made in
	Regions map[string][]types.Rate
}

func LoadConfig(path string) error {
	//TODO
	return nil
}

func ApiKey() string {
	//TODO
	return ""
}

func Regions() map[string][]types.Rate {
	//TODO
	return nil
}
