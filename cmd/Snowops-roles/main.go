package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: failed to load .env file: %v", err)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	address := port
	if !strings.Contains(port, ":") {
		address = ":" + port
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("starting server on port %s", port)

	if err := router.Run(address); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
