package handlers

type responseGameRoom struct {
	Hash    string `json:"hash"`
	Players int    `json:"players"`
}
