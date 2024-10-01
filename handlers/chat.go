package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/markbates/goth/gothic"
)

type Client struct {
	UserID string
	Conn *websocket.Conn
	Send chan []byte
}

type ChatHub struct {
	Clients 					map[*Client]bool
	Broadcast 				chan []byte
	ClientRegister 		chan *Client
	ClientUnregister 	chan *Client
}

var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.ClientRegister:
			h.Clients[client] = true
		case client := <-h.ClientUnregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func ChatRoom(c *gin.Context, hub *ChatHub) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}

	// Retrieve the session to get the user's name
	userID, err := gothic.GetFromSession("userID", c.Request)
	if userID == "" {
		// If user is not authenticated, close the connection
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Unauthorized"))
		conn.Close()
		return
	}
	if err != nil {
		log.Println("Failed to get user name from session:", err)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Internal Server Error"))
		conn.Close()
		return
	}

	client := &Client{Conn: conn, Send: make(chan []byte, 256), UserID: userID}
	hub.ClientRegister <- client

	// Handle concurrent message sending and receiving
	go client.writePump()
	go client.readPump(hub)
}

func (c *Client) readPump(hub *ChatHub) {
	defer func() {
		hub.ClientUnregister <- c
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		hub.Broadcast <- []byte(c.UserID + ": " + string(message))
	}
}

func (c *Client) writePump() {
	for message := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, message)
	}
	// The hub closed the channel
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}