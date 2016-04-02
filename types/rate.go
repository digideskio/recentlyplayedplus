package types

//Region represents a regional endpoint that the Riot API can target
type Rate struct {
	//The number of seconds for which the max requests can occur within
	Period uint32
	//The maximum number of requests allowed in the given period.
	Max uint32
}
