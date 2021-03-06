package handlers

import (
	"net/http"

	cnt "context"

	"github.com/go-park-mail-ru/2018_2_LSP_GAME/user"
	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user_proto"
	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
)

func handleGameRoomConnect(c *websocket.Conn, room *GameRoom, u user.User) error {
	defer deleteGameIfnecessary(room)
	subscription := room.Subscribe()
	defer room.Unsubscribe(subscription)

	room.Join(u)
	defer room.Leave(u)

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
				return nil
			}
		case cmd, ok := <-newCommands:
			if !ok {
				return nil
			}

			room.Execute(u, cmd)
		}
	}
}

// CreateGameRoom create new game room
func CreateGameRoom(env *Env, w http.ResponseWriter, r *http.Request) error {
	title, err := parseRoomTitleFromURL(r)
	players := parsePlayersCountFromURL(r)
	mapSize := parseMapSizeFromURL(r)
	timeLimit := parseTimeLimitFromURL(r)

	claims := context.Get(r, "claims").(map[string]interface{})
	userID := int(claims["id"].(float64))

	ctx := cnt.Background()
	userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
	userGRPC, err := userManager.GetOne(ctx,
		&user_proto.UserID{
			ID: int64(userID),
		})
	if err := handleGetOneUserGrpcError(env, err); err != nil {
		return err
	}

	u := convertGRPCUserToInternal(userGRPC)

	if err := checkUserAlreadyInGame(u); err != nil {
		return err
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	roomHash := generateRoomHash()
	room := NewGameRoom(roomHash, title, players, mapSize, timeLimit)
	rooms[roomHash] = room

	if err = c.WriteJSON(makeEventCustom("room", u, map[string]interface{}{"hash": roomHash})); err != nil {
		return nil
	}

	gameCount.Inc()

	if err := handleGameRoomConnect(c, room, u); err != nil {
		return err
	}

	return nil
}

// GetAllGames returns all available games
func GetAllGames(env *Env, w http.ResponseWriter, r *http.Request) error {
	allgames := make([]responseGameRoom, 0)
	for _, r := range rooms {
		allgames = append(allgames, convertGameRoomToResponse(r))
	}
	return StatusData{
		Code: http.StatusOK,
		Data: map[string][]responseGameRoom{
			"gamerooms": allgames,
		},
	}
}

// ConnectToGameRoom connects to game room
func ConnectToGameRoom(env *Env, w http.ResponseWriter, r *http.Request) error {
	roomHash, err := parseRoomHashFromURL(r)
	if err != nil {
		return err
	}
	if _, ok := rooms[roomHash]; !ok {
		return StatusData{http.StatusNotFound, map[string]string{"error": "Game not found"}}
	}

	claims := context.Get(r, "claims").(map[string]interface{})
	userID := int(claims["id"].(float64))

	ctx := cnt.Background()
	userManager := user_proto.NewUserCheckerClient(env.GRCPUser)
	userGRPC, err := userManager.GetOne(ctx,
		&user_proto.UserID{
			ID: int64(userID),
		})
	if err := handleGetOneUserGrpcError(env, err); err != nil {
		return err
	}

	u := convertGRPCUserToInternal(userGRPC)

	if err := checkUserAlreadyInGame(u); err != nil {
		return err
	}
	if err := checkRoomLimit(rooms[roomHash]); err != nil {
		return err
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close() // nolint: errcheck

	if err := handleGameRoomConnect(c, rooms[roomHash], u); err != nil {
		return err
	}

	return nil
}
