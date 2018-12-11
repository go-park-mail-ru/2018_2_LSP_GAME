package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
)

var roomHashSize = 15

var gameCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "games_total",
		Help: "How many Games are processed.",
	},
)

var rooms = make(map[string]*GameRoom)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	prometheus.MustRegister(gameCount)
}
