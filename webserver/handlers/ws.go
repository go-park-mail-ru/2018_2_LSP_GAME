package handlers

import (
	"math/rand"
	"net/http"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-park-mail-ru/2018_2_LSP_GAME/user"
	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
)

type WSMessage struct {
	command string
	data    string
}

var rooms = make(map[string]*GameRoom)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleGameRoomConnect(c *websocket.Conn, roomHash string, u user.User) error {
	subscription := rooms[roomHash].Subscribe()
	defer rooms[roomHash].Unsubscribe(subscription)

	rooms[roomHash].Join(u)
	defer rooms[roomHash].Leave(u)

	c.WriteJSON(map[string]string{"Type": "room", "Hash": roomHash})

	newCommands := make(chan Command)
	go func() {
		var cmd Command
		for {
			err := c.ReadJSON(&cmd)
			if err != nil {
				close(newCommands)
				return
			}
			newCommands <- cmd
		}
	}()

	for {
		select {
		case event := <-subscription.New:
			if c.WriteJSON(&event) != nil {
				// They disconnected.
				return nil
			}
		case cmd, ok := <-newCommands:
			// If the channel is closed, they disconnected.
			if !ok {
				return nil
			}

			// Otherwise, say something.
			rooms[roomHash].Execute(u, cmd)
		}
	}
}

func userAlreadyInGame(u user.User) bool {
	for i := range rooms {
		if rooms[i].UserIn(u) {
			return true
		}
	}
	return false
}

func CreateGameRoom(env *Env, w http.ResponseWriter, r *http.Request) error {
	claims := context.Get(r, "claims").(jwt.MapClaims)
	userID, _ := strconv.Atoi(claims["id"].(string))
	u, err := user.GetOne(env.DB, userID)
	if err != nil {
		return StatusData{http.StatusInternalServerError, map[string]string{"error": err.Error()}}
	}

	if userAlreadyInGame(u) {
		return StatusData{http.StatusConflict, map[string]string{"error": "User is alredy in game"}}
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	roomHash := RandStringRunes(15)
	_, ok := rooms[roomHash]
	for ok {
		roomHash = RandStringRunes(15)
		_, ok = rooms[roomHash]
	}

	rooms[roomHash] = NewGameRoom()
	go rooms[roomHash].Run()

	handleGameRoomConnect(c, roomHash, u)
	if len(rooms[roomHash].users) == 0 {
		delete(rooms, roomHash)
	}

	return StatusData{http.StatusOK, map[string]string{"error": "err"}}
}

func GetAllGames(env *Env, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func ConnectToGameRoom(env *Env, w http.ResponseWriter, r *http.Request) error {
	roomHashURL, ok := r.URL.Query()["room"]

	if !ok || len(roomHashURL[0]) < 1 {
		return StatusData{http.StatusBadRequest, map[string]string{"error": "You must specify room ID"}}
	}
	roomHash := roomHashURL[0]

	if _, ok := rooms[roomHash]; !ok {
		return StatusData{http.StatusNotFound, map[string]string{"error": "Game not found"}}
	}

	claims := context.Get(r, "claims").(jwt.MapClaims)
	userID, _ := strconv.Atoi(claims["id"].(string))
	u, err := user.GetOne(env.DB, userID)
	if err != nil {
		return StatusData{http.StatusInternalServerError, map[string]string{"error": err.Error()}}
	}

	if userAlreadyInGame(u) {
		return StatusData{http.StatusConflict, map[string]string{"error": "User is alredy in game"}}
	}

	if len(rooms[roomHash].users) == 4 {
		return StatusData{http.StatusUnprocessableEntity, map[string]string{"error": "Too many users in game"}}
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	handleGameRoomConnect(c, roomHash, u)
	if len(rooms[roomHash].users) == 0 {
		delete(rooms, roomHash)
	}

	return StatusData{http.StatusOK, map[string]string{"error": "err"}}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
