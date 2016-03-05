package request

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
	region   string
	endpoint string
}

// GetSummoner retrieves information about the specified summoner, given
// their summoner name and region. An API Key must be configured.
func GetSummoner(region, name, apiKey string) Summoner {
	//TODO
	return Summoner{}
}

// GetMatchlist retrieves a summoner's recent match history, given their region
// and region-unique SummonerID. An API Key must be configured.
func GetMatchlist(region string, summonerid int64, apiKey string) Matchlist {
	//TODO
	return Matchlist{}
}

func (r *request) Do() {

}
