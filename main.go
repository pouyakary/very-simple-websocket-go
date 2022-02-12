package main

// ─── IMPORTS ────────────────────────────────────────────────────────────────────

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// ─── WEBSOCKET CONNECTION ───────────────────────────────────────────────────────

var clients = make(map[*websocket.Conn]bool)
var broadcaster = make(chan []byte)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ─── SOCKET ENTRY ───────────────────────────────────────────────────────────────

func socketHandler(responseWriter http.ResponseWriter, reader *http.Request) {

	// initiating
	conn, err := upgrader.Upgrade(responseWriter, reader, nil)
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	clients[conn] = true

	// The event loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			break
		}

		broadcaster <- message
	}

}

// ─── MESSAGE BROADCASTER ────────────────────────────────────────────────────────

func runBroadcaster() {
	for {
		message := <-broadcaster

		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// ─── MAIN ───────────────────────────────────────────────────────────────────────

func main() {
	http.HandleFunc("/socket", socketHandler)
	go runBroadcaster()
	log.Fatal(http.ListenAndServe("localhost:1234", nil))
}
