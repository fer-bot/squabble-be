package main

import (
	"log"
	"os"

	"squabble/controller"
	"squabble/db"
	"squabble/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(os.Getenv("ENV") + ".env")
	if err != nil {
		log.Fatalf("error: failed to load the env file. Err: %s", err)
	}
	db.InitDB()
	db.InitRedis(1)

	models.AutoMigrate()
	go models.GetHub().Run()

	router := gin.New()

	router.POST("/user/login", func(c *gin.Context) {
		controller.LoginHandler(c.Writer, c.Request)
	})
	router.POST("/user/register", func(c *gin.Context) {
		controller.RegisterHandler(c.Writer, c.Request)
	})
	router.GET("/user/logout", func(c *gin.Context) {
		controller.LogoutHandler(c.Writer, c.Request)
	})
	router.POST("/start/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		controller.Start(c.Writer, c.Request, roomId)
	})
	router.POST("/answer/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		controller.Answer(c.Writer, c.Request, roomId)
	})
	router.GET("/listen-game-state/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		sessionId := c.Query("session-id")
		controller.ListenGameWS(c.Writer, c.Request, roomId, sessionId)
	})

	log.Printf("Starting server at Port: " + os.Getenv("PORT"))
	router.Run("0.0.0.0:" + os.Getenv("PORT"))
}
