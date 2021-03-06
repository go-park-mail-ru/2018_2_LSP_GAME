package game

import (
	"errors"
	"sync"
	"time"
)

// Event implements game event
type Event struct {
	Event string
	Data  map[string]interface{}
}

func makeEvent(event string, data map[string]interface{}) Event {
	return Event{event, data}
}

// Game implements main game struct
type Game struct {
	Gamemap        gameMap
	currentPlayer  int
	totalGoldCount int
	players        []player
	Events         chan Event
	moveMutex      *sync.Mutex
	TimeLimit      int
	Timer          *time.Timer
}

// MakeGame is constructor for Game struct
func MakeGame(distribution []Distribution, playersCount int, pirateCount int, timeLimit int) Game {
	game := Game{}
	game.Events = make(chan Event, 10)
	builder := makeMapBuilder(distribution)
	game.Gamemap = builder.generateMap()

	game.currentPlayer = 0
	game.players = make([]player, playersCount)
	for i := 0; i < playersCount; i++ {
		game.players[i].addPirates(pirateCount, position{true, 0})
	}

	game.TimeLimit = timeLimit

	game.moveMutex = &sync.Mutex{}

	return game
}

func (g *Game) StartTimer() {
	g.Timer = time.NewTimer(time.Duration(g.TimeLimit) * time.Second)
	go func() {
		<-g.Timer.C
		g.Events <- makeEvent("expired", map[string]interface{}{
			"playerID": g.currentPlayer,
		})
		g.nextPlayer()
	}()
}

func (g *Game) stopTimer() {
	g.Timer.Stop()
}

func (g *Game) restartTimer() {
	g.Timer.Stop()
}

// RemovePlayer removes player from game on disconnect
func (g *Game) RemovePlayer(playerID int) {
	g.moveMutex.Lock()
	g.players = append(g.players[:playerID], g.players[playerID:]...)
	g.moveMutex.Unlock()
}

func (g *Game) checkForWin() int {
	if g.Gamemap.getUnusedGoldCount() > 0 {
		return -1
	}

	id := 0
	score := g.players[0].getScore()
	for i := range g.players {
		if g.players[i].getScore() > score {
			score = g.players[i].getScore()
			id = i
		}
	}

	return id
}

// GetCurrentPlayerID returns current player ID
func (g *Game) GetCurrentPlayerID() int {
	return g.currentPlayer
}

func (g *Game) getCurrentPlayer() player {
	return g.players[g.currentPlayer]
}

func (g *Game) nextPlayer() {
	// Меняем ход игрока
	g.currentPlayer = (g.currentPlayer + 1) % len(g.players)
	g.Events <- makeEvent("nextplayer", map[string]interface{}{
		"playerID": g.currentPlayer,
	})
	g.StartTimer()
}

// MovePirate moves pirate to cardID (if it is possible)
func (g *Game) MovePirate(pirateID int, cardID int) error {
	g.stopTimer()

	g.moveMutex.Lock()
	defer g.moveMutex.Unlock()

	player := g.getCurrentPlayer()
	pirate := player.getPirate(pirateID)
	pirateCard := pirate.getCard()

	moveableCards := g.Gamemap.getMoveableCards(g.currentPlayer, pirateCard)
	movable := false
	for i := 0; i < len(moveableCards); i++ {
		if moveableCards[i].id == cardID {
			movable = true
			break
		}
	}
	if !movable {
		return nil
	}

	// Проверяем всех остальных пиратов. Если они стоят на этой карточке - их нужно убить
	for i := 0; i < len(g.players); i++ {
		pirates := g.players[i].getPirates()
		for j := 0; j < len(pirates); j++ {
			if pirates[j].getCard().id == cardID {
				g.players[i].pirates[j].kill()
				g.Events <- makeEvent("kill", map[string]interface{}{
					"playerID": i,
					"pirateID": j,
				})
				break
			}
		}
	}

	// Перемещаем пирата
	g.players[g.currentPlayer].pirates[pirateID].setCard(makePosition(false, cardID))

	// Применяем карточку
	cardType := g.Gamemap.getCardType(cardID)

	g.Events <- makeEvent("movement", map[string]interface{}{
		"playerID": g.currentPlayer,
		"pirateID": pirateID,
		"cardID":   cardID,
		"cardType": cardType,
	})

	cardObject := makeCard(cardType, pirateID)
	cardObject.apply(g)

	// Проверяем условие победы
	if id := g.checkForWin(); id != -1 {
		g.Events <- makeEvent("win", map[string]interface{}{
			"playerID": id,
		})
		return errors.New("Game ended")
	}

	g.nextPlayer()

	return nil

}
