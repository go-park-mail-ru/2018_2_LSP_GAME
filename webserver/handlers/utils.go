package handlers

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2018_2_LSP_GAME/user"
	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user_proto"
)

func convertGRPCUserToInternal(u *user_proto.User) user.User {
	converted := user.User{
		ID:        int(u.ID),
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
	}
	return converted
}

func randStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func generateRoomHash() string {
	roomHash := randStringRunes(roomHashSize)
	_, ok := rooms[roomHash]
	for ok {
		roomHash = randStringRunes(roomHashSize)
		_, ok = rooms[roomHash]
	}
	return roomHash
}

func checkRoomLimit(room *GameRoom) error {
	if len(room.users) == room.MaxUsers {
		return StatusData{
			Code: http.StatusUnprocessableEntity,
			Data: map[string]string{
				"error": "Too many users in game",
			},
		}
	}
	return nil
}

func checkUserAlreadyInGame(u user.User) error {
	for i := range rooms {
		if rooms[i].UserIn(u) {
			return StatusData{
				Code: http.StatusConflict,
				Data: map[string]string{
					"error": "User is alredy in game",
				},
			}
		}
	}
	return nil
}

func parseRoomHashFromURL(r *http.Request) (string, error) {
	roomHashURL, ok := r.URL.Query()["room"]
	if !ok || len(roomHashURL[0]) < 1 {
		return "", StatusData{
			Code: http.StatusBadRequest,
			Data: map[string]string{
				"error": "You must specify room ID",
			},
		}
	}
	return roomHashURL[0], nil
}

func parseRoomTitleFromURL(r *http.Request) (string, error) {
	roomTitleURL, ok := r.URL.Query()["title"]
	if !ok || len(roomTitleURL[0]) < 1 {
		return "", StatusData{
			Code: http.StatusBadRequest,
			Data: map[string]string{
				"error": "No title was specified",
			},
		}
	}
	return roomTitleURL[0], nil
}

func parsePlayersCountFromURL(r *http.Request) int {
	parsedURL, ok := r.URL.Query()["players"]
	if !ok {
		return 4
	}
	cnt, err := strconv.Atoi(parsedURL[0])
	if err == nil && cnt == 2 {
		return 2
	}
	return 4
}

func parseMapSizeFromURL(r *http.Request) int {
	parsedURL, ok := r.URL.Query()["size"]
	if !ok {
		return 5
	}
	cnt, err := strconv.Atoi(parsedURL[0])
	if err == nil && (cnt == 7 || cnt == 9) {
		return cnt
	}
	return 5
}

func parseTimeLimitFromURL(r *http.Request) int {
	parsedURL, ok := r.URL.Query()["timelimit"]
	if !ok {
		return 60
	}
	cnt, err := strconv.Atoi(parsedURL[0])
	if err == nil && (cnt == 30 || cnt == 90) {
		return cnt
	}
	return 60
}

func deleteGameIfnecessary(room *GameRoom) {
	if len(room.users) == 0 {
		gameCount.Dec()
		delete(rooms, room.Hash)
	}
}

func convertGameRoomToResponse(gr *GameRoom) responseGameRoom {
	res := responseGameRoom{
		Hash:       gr.Hash,
		Players:    len(gr.users),
		Title:      gr.Title,
		MaxPlayers: gr.MaxUsers,
		TimeLimit:  gr.game.TimeLimit,
		MapSize:    gr.game.Gamemap.Size,
	}
	return res
}
