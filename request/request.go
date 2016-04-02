package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
}

// Request contains information about an HTTP request to make to the Riot API.
// Executing a request will queue it against the respective development key's
// request rate, making sure it does not exceed the rate.
// Note that all Riot API endpoints respond only to GET requests, and therefore
// tracking of the request method is not necessary.
type request struct {
	url  string
	body chan []byte
	err  chan error
}

// GetSummoners retrieves information about the specified summoners, given
// their summoner name and region. An API Key must be configured.
func GetSummoners(region string, names ...string) (types.Summoner, error) {
	endpoint := fmt.Sprintf("/api/lol/%s/v1.4/summoner/by-name/%s", region, strings.Join(names, ", "))
	req := getBaseRequest(region, endpoint)
	//Throwing away queue position for now. Can be used for logging later.
	_, err := lim.Enqueue(req, region)
	if err != nil {
		return types.Summoner{}, err
	}
	var response []byte
	select {
	case response = <-req.body:
	case err = <-req.err:
		return types.Summoner{}, err
	}
	ret := types.Summoner{}
	json.Unmarshal(response, &ret)
	return ret, nil
}

// GetRecentGames retrieves a summoner's recent match history, given their region
// and region-unique SummonerID. An API Key must be configured.
func GetRecentGames(region string, summonerid int64, apiKey string) (types.Matchlist, error) {
	endpoint := fmt.Sprintf("/api/lol/%s/v1.3/game/by-summoner/%d", region, summonerid)
	req := getBaseRequest(region, endpoint)
	//Throwing away queue position for now. Can be used for logging later.
	_, err := lim.Enqueue(req, region)
	if err != nil {
		return types.Matchlist{}, err
	}
	var response []byte
	select {
	case response = <-req.body:
	case err = <-req.err:
		return types.Matchlist{}, err
	}
	ret := types.Matchlist{}
	json.Unmarshal(response, &ret)
	return ret, nil
}

func (r request) Do() {
	resp, err := http.Get(r.url)
	if err != nil {
		r.err <- err
		return
	}
	if resp.StatusCode/100 != 2 {
		r.err <- fmt.Errorf("Request returned '%s'", resp.Status)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.err <- err
		return
	}
	r.body <- body
}

func getBaseURL(region string) string {
	return fmt.Sprintf("https://%[1]s.api.pvp.net/api/lol/%[1]s", region)
}

func glueURL(base, endpoint, devKey string) string {
	return fmt.Sprintf("%s%s?api_key=%s", base, endpoint, devKey)
}

func getBaseRequest(region, endpoint string) request {
	return request{
		url:  glueURL(getBaseURL(region), endpoint, config.ApiKey()),
		body: make(chan []byte, 1),
		err:  make(chan error, 1),
	}
}
