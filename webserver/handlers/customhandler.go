package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StatusData represents an error with an associated HTTP status code.
type StatusData struct {
	Code int
	Data interface{}
}

// Error allows StatusData to satisfy the error interface.
func (sd StatusData) Error() string {
	return fmt.Sprintf("%v", sd.Data)
}

// GetJSONData allows StatusData to satisfy the error interface.
func (sd StatusData) GetJSONData() ([]byte, error) {
	return json.Marshal(sd.Data)
}

// Status returns our HTTP status code.
func (se StatusData) Status() int {
	return se.Code
}

// Env hold env
type Env struct {
	Logger   *zap.SugaredLogger
	GRCPUser *grpc.ClientConn
	GRCPAuth *grpc.ClientConn
}

// HandlerFunc func for Handler
type HandlerFunc func(e *Env, w http.ResponseWriter, r *http.Request) error

// HandlersMap map of handlers
type HandlersMap map[string]HandlerFunc

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type Handler struct {
	*Env
	H HandlerFunc
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case StatusData:
			w.WriteHeader(e.Status())
			jsonData, err := e.GetJSONData()
			if err != nil {
				h.Logger.Errorw("Can't get JSON data from StatusCode",
					"error", e,
				)
				return
			}
			_, err = w.Write(jsonData)
			if err != nil {
				h.Logger.Errorw("Can't write JSON data to response body",
					"data", jsonData,
				)
				return
			}
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}
