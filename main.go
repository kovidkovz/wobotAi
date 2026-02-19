package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrader helps us turn a normal HTTP connection into a WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Letting everyone in for now!
	},
}

func main() {
	// Setting up our router using Gin
	r := gin.Default()

	hub := newHub()
	go hub.run()

	// Just a quick check to see if the server is alive
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Let's see who is currently connected
	r.GET("/clients", func(c *gin.Context) {
		ids := hub.getClients()
		c.JSON(200, gin.H{
			"clients": ids,
			"count":   len(ids),
		})
	})

	// This is where the WebSocket magic happens
	r.GET("/ws", func(c *gin.Context) {
		serveWs(hub, c)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Starting the server in the background so it doesn't block other important things
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Waiting for a signal to stop the server politely
	quit := make(chan os.Signal, 1)
	// Catching different kill signals to ensure we clean up nicely
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Giving the server 5 seconds to finish what it's doing before shutting down completely
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
