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

	usersRoutes := router.Group("/users")
	usersRoutes.GET("", controllers.GetUsers)
	usersRoutes.POST("", controllers.CreateUser)
	usersRoutes.PUT("/:id", controllers.UpdateUser)
	usersRoutes.DELETE("/:id", controllers.DeleteUser)
	usersRoutes.POST("/login", controllers.Login)

	authRoutes := router.Group("")
	authRoutes.Use(controllers.AuthMiddleware())
	// group untuk mengakses fungsi yang perlu login
	// dalam hal ini, hanya untuk mengakses fungsi logout

	authRoutes.POST("/users/logout", controllers.Logout)

	if err := router.Run(":" + port); err != nil {
		panic(err)
	}
}
