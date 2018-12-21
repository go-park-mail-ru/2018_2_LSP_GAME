package game

const (
	DefaultCard = iota
	GoldCard    = iota
	KillCard    = iota
)

type card interface {
	apply(game *Game)
}

type emptyCard struct {
}

func (c emptyCard) apply(game *Game) {
}

type goldCard struct {
	pirateID int
}

func (c goldCard) apply(game *Game) {
	if game.Gamemap.goldData[game.players[game.currentPlayer].pirates[c.pirateID].card.id] > 0 {
		game.players[game.currentPlayer].incScore()
		game.Gamemap.goldData[game.players[game.currentPlayer].pirates[c.pirateID].card.id]--
		game.Events <- makeEvent("card_gold", map[string]interface{}{
			"playerID": game.currentPlayer,
		})
	}
}

type killCard struct {
	pirateID int
}

func (c killCard) apply(game *Game) {
	game.players[game.currentPlayer].pirates[c.pirateID].setCard(makePosition(true, 0))
	game.Events <- makeEvent("card_kill", map[string]interface{}{
		"playerID": game.currentPlayer,
		"pirateID": c.pirateID},
	)
}

func makeCard(id int, pirateID int) card {
	switch id {
	case GoldCard:
		return goldCard{pirateID}
	case KillCard:
		return killCard{pirateID}
	default:
		return emptyCard{}
	}
}
