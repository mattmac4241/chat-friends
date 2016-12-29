package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mattmac4241/chat-auth/service"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dburl := os.Getenv("DBURL")

	dbinfo := fmt.Sprintf("%s", dburl)
	db, err := service.InitDatabase(dbinfo)
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redis, err := service.InitRedisClient(redisAddress, redisPassword)
	if err != nil {
		log.Fatal("Failed to connect to redis")
	}

	service.REDIS = redis
	service.DB = db

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3001"
	}
	server := service.NewServer()
	server.Run(":" + port)
}
