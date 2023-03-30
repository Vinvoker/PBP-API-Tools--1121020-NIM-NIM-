package main

import (
	"gin/controllers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	router := gin.Default()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port := os.Getenv("ROUTER_PORT")

	router.POST("/cron", controllers.ActivateCRON)
	router.POST("/email", controllers.SendEmail)

	if err := router.Run(":" + port); err != nil {
		panic(err)
	}
}
