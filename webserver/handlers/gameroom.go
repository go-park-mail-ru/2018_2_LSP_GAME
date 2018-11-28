package handlers

import (
	"container/list"
	"strconv"
	"time"

	game "github.com/go-park-mail-ru/2018_2_LSP_GAME/game"
	"github.com/go-park-mail-ru/2018_2_LSP_GAME/user"
)

type Command struct {
	Action string            `json:"action"`
	Params map[string]string `json:"params"`
}

type GameRoom struct {
	subscribe   chan (chan<- Subscription)
	unsubscribe chan (<-chan Event)
	publish     chan Event
	users       []user.User
	game        game.Game
	Hash        string
}

// NewGameRoom creates new game room
func NewGameRoom(hash string) *GameRoom {
	g := MakeGameRoom(hash)
	go g.Run()
	return &g
}

// MakeGameRoom makes new game room
func MakeGameRoom(hash string) GameRoom {
	c := GameRoom{
		subscribe:   make(chan (chan<- Subscription), 10),
		unsubscribe: make(chan (<-chan Event), 10),
		publish:     make(chan Event, 10),
		Hash:        hash,
	}

	distribution := []game.Distribution{game.MakeDistribution(game.DefaultCard, 12), game.MakeDistribution(game.GoldCard, 8), game.MakeDistribution(game.KillCard, 5)}

	c.game = game.MakeGame(distribution, 2, 2)
	return c
}

// Event game room event
type Event struct {
	Type string
	User struct {
		Username string
		ID       int
	}
	Timestamp int
	Data      map[string]interface{}
}

func makeEventCustom(typ string, u user.User, data map[string]interface{}) Event {
	event := Event{
		Type:      typ,
		Timestamp: int(time.Now().Unix()),
		Data:      data,
	}
	event.User.ID = u.ID
	event.User.Username = u.Username
	return event
}

func makeEventMessage(u user.User, msg string) Event {
	event := Event{
		Type:      "message",
		Timestamp: int(time.Now().Unix()),
		Data:      map[string]interface{}{"message": msg},
	}
	event.User.ID = u.ID
	event.User.Username = u.Username
	return event
}

func makeEventFromGame(gameevent game.Event) Event {
	event := Event{
		Type:      "game",
		Timestamp: int(time.Now().Unix()),
		Data:      gameevent.Data,
	}
	return event
}

// Subscription implement subscription
type Subscription struct {
	Archive []Event
	New     <-chan Event
}

// Unsubscribe unsubscribes user
func (gr *GameRoom) Unsubscribe(s Subscription) {
	gr.unsubscribe <- s.New
	drain(s.New)
}

// Subscribe subscribes user
func (gr *GameRoom) Subscribe() Subscription {
	resp := make(chan Subscription)
	gr.subscribe <- resp
	return <-resp
}

func (gr *GameRoom) UserIn(u user.User) bool {
	for i := range gr.users {
		if gr.users[i].ID == u.ID {
			return true
		}
	}
	return false
}

// Join event spawn
func (gr *GameRoom) Join(u user.User) {
	gr.users = append(gr.users, u)
	gr.publish <- makeEventCustom("join", u, map[string]interface{}{})
}

// Execute user command
func (gr *GameRoom) Execute(u user.User, cmd Command) {
	switch cmd.Action {
	case "move":
		id := -1
		for i := range gr.users {
			if gr.users[i].ID == u.ID {
				id = i
				break
			}
		}
		if id != gr.game.GetCurrentPlayerID() {
			return
		}
		pirate, err := strconv.Atoi(cmd.Params["pirate"])
		if err != nil {
			return
		}
		card, err := strconv.Atoi(cmd.Params["card"])
		if err != nil {
			return
		}
		err = gr.game.MovePirate(pirate, card)
		if err != nil {
			return
		}
	default:
		return
	}
}

// Leave game room
func (gr *GameRoom) Leave(u user.User) {
	id := -1
	for i := 0; i < len(gr.users); i++ {
		if gr.users[i].ID == u.ID {
			gr.users = append(gr.users[:i], gr.users[i+1:]...)
			id = i
			break
		}
	}
	gr.game.RemovePlayer(id)
	gr.publish <- makeEventCustom("leave", u, map[string]interface{}{})
}

const archiveSize = 10

// Run game room
func (gr *GameRoom) Run() {
	archive := list.New()
	subscribers := list.New()

	for {
		select {
		case ch := <-gr.subscribe:
			var events []Event
			for e := archive.Front(); e != nil; e = e.Next() {
				events = append(events, e.Value.(Event))
			}
			subscriber := make(chan Event, 10)
			subscribers.PushBack(subscriber)
			ch <- Subscription{events, subscriber}

		case event := <-gr.publish:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}
			if archive.Len() >= archiveSize {
				archive.Remove(archive.Front())
			}
			archive.PushBack(event)

		case unsub := <-gr.unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Event) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		case gameevent := <-gr.game.Events:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- makeEventFromGame(gameevent)
			}
		}
	}
}

func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
