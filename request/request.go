package request

import (
	"fmt"
	"strings"

	"github.com/thomasmmitchell/recentlyplayedplus/config"
	"github.com/thomasmmitchell/recentlyplayedplus/types"
)

// LimitedDoer is an object representing a task to be performed asynchronously,
// at a time which is managed by a Limiter object.
// When the Limiter deems that there is allowance for another task to be performed,
// this interface's Do() method is called.
type LimitedDoer interface {
	//Performs the task that this object... does. Called by the limiter when there
	// is allowance for the task to be done in the given region.
	Do()
}

//A limiter class to be used with the Riot API requests.
var lim *Limiter

func init() {
	lim = NewLimiter()
	//TODO: Set up regions and rates
}

// Request contains information about an HTTP request to make to the Riot API.
// Executing a request will queue it against the respective development key's
// request rate, making sure it does not exceed the rate.
// Note that all Riot API endpoints respond only to GET requests, and therefore
// tracking of the request method is not necessary.
type request struct {
	url  string
	body chan string
	err  chan error
}

// GetSummoners retrieves information about the specified summoners, given
// their summoner name and region. An API Key must be configured.
func GetSummoners(region string, names ...string) types.Summoner {
	endpoint := fmt.Sprintf("/api/lol/%s/v1.4/summoner/by-name/%s", region, strings.Join(names, ", "))
	req := request{
		url:  glueURL(getBaseURL(region), endpoint, config.APIKey),
		body: make(chan string),
		err:  make(chan error),
	}
	lim.Enqueue(req, region)
	//TODO: Read the result, unmarshal it
	return types.Summoner{}
}

// GetMatchlist retrieves a summoner's recent match history, given their region
// and region-unique SummonerID. An API Key must be configured.
func GetMatchlist(region string, summonerid int64, apiKey string) types.Matchlist {
	//TODO
	return types.Matchlist{}
}

func (r request) Do() {

}

func getBaseURL(region string) string {
	return fmt.Sprintf("https://%[1]s.api.pvp.net/api/lol/%[1]s", region)
}

func glueURL(base, endpoint, devKey string) string {
	return fmt.Sprintf("%s%s?api_key=%s", base, endpoint, devKey)
}
