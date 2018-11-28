package routes

import (
	"net/http"

	"github.com/go-park-mail-ru/2018_2_LSP_GAME/webserver/handlers"
	"github.com/go-park-mail-ru/2018_2_LSP_GAME/webserver/middlewares"
)

func Get() handlers.HandlersMap {
	handlersMap := handlers.HandlersMap{}
	handlersMap["/game"] = makeRequest(handlers.HandlersMap{
		"get": middlewares.Auth(handlers.ConnectToGameRoom),
	})
	handlersMap["/games"] = makeRequest(handlers.HandlersMap{
		"get": middlewares.Auth(handlers.GetAllGames),
	})
	handlersMap["/gamecreate"] = makeRequest(handlers.HandlersMap{
		"get": middlewares.Auth(handlers.CreateGameRoom),
	})
	return handlersMap
}

func makeRequest(handlersMap handlers.HandlersMap) handlers.HandlerFunc {
	return func(env *handlers.Env, w http.ResponseWriter, r *http.Request) error {
		var key string
		switch r.Method {
		case http.MethodGet:
			key = "get"
		case http.MethodPost:
			key = "post"
		case http.MethodPut:
			key = "put"
		case http.MethodDelete:
			key = "delete"
		default:
			return middlewares.Cors(handlers.DefaultHandler)(env, w, r)
		}
		if _, ok := handlersMap[key]; ok {
			return middlewares.Cors(handlersMap[key])(env, w, r)
		}
		return middlewares.Cors(handlers.DefaultHandler)(env, w, r)
	}
}
