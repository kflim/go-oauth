package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	"github.com/kflim/go-oauth/handlers"
)

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	clientCallbackURL := os.Getenv("CLIENT_CALLBACK_URL")
	sessionKey := os.Getenv("SESSION_KEY")

	if sessionKey == "" {
		sessionKey = "default_session_key" // Replace with a secure key in production
	}

	if clientID == "" || clientSecret == "" || clientCallbackURL == "" {
		log.Fatal("Environment variables (CLIENT_ID, CLIENT_SECRET, CLIENT_CALLBACK_URL) are required")
	}

	goth.UseProviders(
		google.New(clientID, clientSecret, clientCallbackURL),
	)

	store := sessions.NewCookieStore([]byte(sessionKey))
	gothic.Store = store

	var hub = handlers.ChatHub {
		Clients:    make(map[*handlers.Client]bool),
		Broadcast:  make(chan []byte),
		ClientRegister:   make(chan *handlers.Client),
		ClientUnregister : make(chan *handlers.Client),
	}

	go hub.Run()

	r.LoadHTMLGlob("templates/*")

	r.GET("/", handlers.Home)
	r.GET("/auth/:provider", handlers.SignInWithProvider)
	r.GET("/auth/:provider/callback", handlers.CallbackHandler)

	r.GET("/ws", func(c *gin.Context) {
    // Assuming you have a global or accessible hub instance
    handlers.ChatRoom(c, &hub)
	})

	r.GET("/success", handlers.Success)
	r.GET("/retry-login", handlers.RetryLogin)

	r.Run(":5000")
}