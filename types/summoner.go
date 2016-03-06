package types

//Summoner (account) for League of Legends.
type Summoner struct {
	Name   string `json:"name"` //Username string of the player.
	ID     uint64 `json:"id"`   //Unique ID (within region) of the player.
	Region string
}
