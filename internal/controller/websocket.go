package controller

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	MessageTypeConnect     MessageType = "connect"
	MessageTypeCurrentSong MessageType = "current_song"
	MessageTypeQueue       MessageType = "queue"
)

type CurrentSongPayload struct {
	Artist    string `json:"artist"`
	Title     string `json:"title"`
	ArtUrl    string `json:"art_url"`
	Duration  int    `json:"duration"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	ID        string `json:"id"`
}

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type WebsocketController struct {
	clients         map[*websocket.Conn]bool
	sendOnNewClient *Message
}

var upgrader = websocket.Upgrader{}

func NewWebsocketController() *WebsocketController {
	return &WebsocketController{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (wsc *WebsocketController) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	wsc.clients[conn] = true
	return conn, nil
}

func (wsc *WebsocketController) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("websocket connection")

		conn, err := wsc.Upgrade(w, r)
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
		}

		wsc.clients[conn] = true

		if wsc.sendOnNewClient != nil {
			conn.WriteJSON(wsc.sendOnNewClient)
		}
	})

	fmt.Println("websocket routes registered")
}

func (wsc *WebsocketController) Broadcast(message *Message) {
	for client := range wsc.clients {
		err := client.WriteJSON(message)
		if err != nil {
			client.Close()
			delete(wsc.clients, client)
		}
	}
}

func (wsc *WebsocketController) BroadcastOnNewClient(message *Message) {
	wsc.sendOnNewClient =  message
}
