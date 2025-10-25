package main

import (
	"log"
	"net/http"
	"time"

	"social-network/backend/bus"
	"social-network/backend/db"
	"social-network/backend/handlers"
	"social-network/backend/utils"
	"strconv"

	"github.com/rs/cors"
)

func main() {
	db.InitDB() // connect + run migrations
	// inject DB into utils package for session helpers
	utils.SetDB(db.DB)

	mux := http.NewServeMux()
	RegisterRoutes(mux)

	// CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:5174","http://social-network-frontend"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)

	// Start periodic session cleanup
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			handlers.CleanupSessions()
		}
	}()

	// Start bus forwarder: listen for notification messages and send to WS clients
	go func() {
		for nm := range bus.NotificationChan {
			// if connected, push payload
			sid := strconv.FormatInt(nm.RecipientID, 10)
			if client, ok := clients[sid]; ok {
				client.Send <- nm.Payload
			}
		}
	}()

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
