package request

//Summoner (account) for League of Legends.
type Summoner struct {
	Name   string //Username string of the player.
	ID     uint64 //Unique ID (within region) of the player.
	Region string //Region for which the player is registered.
}
