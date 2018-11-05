package handlers

import (
	"container/list"
	"time"

	"github.com/go-park-mail-ru/2018_2_LSP_GAME/user"
)

type Command struct {
	action string
	params map[string]string
}

type GameRoom struct {
	// Send a channel here to get room events back.  It will send the entire
	// archive initially, and then new messages as they come in.
	subscribe chan (chan<- Subscription)
	// Send a channel here to unsubscribe.
	unsubscribe chan (<-chan Event)
	// Send events here to publish them.
	publish chan Event
	users   []user.User
}

func NewGameRoom() *GameRoom {
	c := new(GameRoom)
	c.subscribe = make(chan (chan<- Subscription), 10)
	c.unsubscribe = make(chan (<-chan Event), 10)
	c.publish = make(chan Event, 10)
	return c
}

func MakeGameRoom() GameRoom {
	c := GameRoom{}
	c.subscribe = make(chan (chan<- Subscription), 10)
	c.unsubscribe = make(chan (<-chan Event), 10)
	c.publish = make(chan Event, 10)
	return c
}

type Event struct {
	Type      string // "join", "leave", or "message"
	User      string
	Timestamp int    // Unix timestamp (secs)
	Text      string // What the user said (if Type == "message")
}

type Subscription struct {
	Archive []Event      // All the events from the archive.
	New     <-chan Event // New events coming in.
}

func (game *GameRoom) Unsubscribe(s Subscription) {
	game.unsubscribe <- s.New
	drain(s.New)
}

func newEvent(typ, user, msg string) Event {
	return Event{typ, user, int(time.Now().Unix()), msg}
}

func (game *GameRoom) Subscribe() Subscription {
	resp := make(chan Subscription)
	game.subscribe <- resp
	return <-resp
}

func (game *GameRoom) UserIn(u user.User) bool {
	for i := range game.users {
		if game.users[i].ID == u.ID {
			return true
		}
	}
	return false
}

func (game *GameRoom) Join(u user.User) {
	game.users = append(game.users)
	game.publish <- newEvent("join", u.Username, "")
}

func (game *GameRoom) Execute(user user.User, cmd Command) {
	// Тут обрабатываем команду
	// game.publish <- newEvent("message", user, message)
}

func (game *GameRoom) Leave(u user.User) {
	game.publish <- newEvent("leave", u.Username, "")
}

const archiveSize = 10

// This function loops forever, handling the game room pubsub
func (game *GameRoom) Run() {
	archive := list.New()
	subscribers := list.New()

	for {
		select {
		case ch := <-game.subscribe:
			var events []Event
			for e := archive.Front(); e != nil; e = e.Next() {
				events = append(events, e.Value.(Event))
			}
			subscriber := make(chan Event, 10)
			subscribers.PushBack(subscriber)
			ch <- Subscription{events, subscriber}

		case event := <-game.publish:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}
			if archive.Len() >= archiveSize {
				archive.Remove(archive.Front())
			}
			archive.PushBack(event)

		case unsub := <-game.unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Event) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		}
	}
}

// func init() {
// 	go Gameroom()
// }

// Helpers

// Drains a given channel of any messages.
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