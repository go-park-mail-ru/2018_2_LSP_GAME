package handlers

type responseGameRoom struct {
	Hash       string `json:"hash"`
	Players    int    `json:"players"`
	Title      string `json:"title"`
	MaxPlayers int    `json:"maxplayers"`
	TimeLimit  int    `json:"timelimit"`
	MapSize    int    `json:"mapsize"`
}
