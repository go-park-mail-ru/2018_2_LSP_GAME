package game

import (
	"fmt"
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
	gamemap        gameMap
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
	game.gamemap = builder.generateMap()

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

func (g *Game) checkForWin() bool {
	return false
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
}

// MovePirate moves pirate to cardID (if it is possible)
func (g *Game) MovePirate(pirateID int, cardID int) error {
	// fmt.Println("Current player", g.currentPlayer)
	// fmt.Println("Players:")
	// for i := 0; i < len(g.players); i++ {
	// 	fmt.Println("Player", i)
	// 	fmt.Println("Score:", g.players[i].score)
	// 	for j := 0; j < len(g.players[i].pirates); j++ {
	// 		fmt.Println("Pirate:", j, g.players[i].pirates[j])
	// 	}
	// 	fmt.Println()
	// }
	// fmt.Println("Map:")
	// for i := 0; i < g.gamemap.size*g.gamemap.size; i++ {
	// 	fmt.Print(g.gamemap.mapData[i], "\t")
	// 	if i%g.gamemap.size == g.gamemap.size-1 {
	// 		fmt.Println()
	// 	}
	// }

	g.stopTimer()
	defer g.StartTimer()

	g.moveMutex.Lock()
	defer g.moveMutex.Unlock()

	player := g.getCurrentPlayer()
	pirate := player.getPirate(pirateID)
	pirateCard := pirate.getCard()

	moveableCards := g.gamemap.getMoveableCards(g.currentPlayer, pirateCard)
	fmt.Println("Movable cards:", moveableCards)
	movable := false
	for i := 0; i < len(moveableCards); i++ {
		if moveableCards[i].id == cardID {
			movable = true
			break
		}
	}
	if !movable {
		fmt.Println("Not movable")
		return nil
	}

	fmt.Println("Movable")

	// Проверяем всех остальных пиратов. Если они стоят на этой карточке - их нужно убить
	for i := 0; i < len(g.players); i++ {
		pirates := g.players[i].getPirates()
		for j := 0; j < len(pirates); j++ {
			if pirates[j].getCard().id == cardID {
				fmt.Println("Убита фишка ", g.players[i].pirates[j], "игрока", g.players[i])
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
	g.Events <- makeEvent("move", map[string]interface{}{
		"playerID": g.currentPlayer,
		"pirateID": pirateID,
	})

	// Применяем карточку
	cardType := g.gamemap.getCardType(cardID)

	g.Events <- makeEvent("card", map[string]interface{}{
		"type": cardType,
	})

	cardObject := makeCard(cardType, pirateID)
	cardObject.apply(g)

	// Проверяем условие победы
	if g.checkForWin() {
		g.Events <- makeEvent("win", map[string]interface{}{
			"playerID": g.currentPlayer,
		})
	}

	g.nextPlayer()

	return nil

}
