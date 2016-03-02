package request

//Matchlist returned by the Riot API. This struct only captures a subset of the
// fields returned by the entire API call - the ones relevant to this purpose.
type Matchlist struct {
	Games []struct {
		FellowPlayers []struct {
			ChampionID int `json:"championId"`
			TeamID     int `json:"teamId"`
			SummonerID int `json:"summonerId"`
		} `json:"fellowPlayers"`
		GameType string `json:"gameType"`
		Stats    struct {
			Win        bool `json:"win"`
			TimePlayed int  `json:"timePlayed"`
		} `json:"stats"`
		GameID     int    `json:"gameId"`
		TeamID     int    `json:"teamId"`
		GameMode   string `json:"gameMode"`
		Invalid    bool   `json:"invalid"`
		SubType    string `json:"subType"`
		CreateDate int64  `json:"createDate"`
	} `json:"games"`
	SummonerID int `json:"summonerId"`
}
