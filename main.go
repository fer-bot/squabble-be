package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"squabble/controller"
	"squabble/db"
	"squabble/form"
	"squabble/models"

	"github.com/joho/godotenv"
)

func SessionAuthMiddleware(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionUUID := r.Header.Get("session-id")
		username, err := models.GetUserModel().VerifyToken(sessionUUID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
			return
		}
		fn(w, r, username)
	}
}

func main() {
	err := godotenv.Load(os.Getenv("ENV") + ".env")
	if err != nil {
		log.Fatalf("error: failed to load the env file. Err: %s", err)
	}
	db.InitDB()
	db.InitRedis(1)

	models.AutoMigrate()
	go models.GetHub().Run()

	http.HandleFunc("/user/login", controller.LoginHandler)
	http.HandleFunc("/user/register", controller.RegisterHandler)
	http.HandleFunc("/user/logout", SessionAuthMiddleware(controller.LogoutHandler))
	http.HandleFunc("/start/", SessionAuthMiddleware(controller.Start))
	http.HandleFunc("/answer/", SessionAuthMiddleware(controller.Answer))
	http.HandleFunc("/listen-game-state/", SessionAuthMiddleware(controller.ListenGameWS))

	log.Printf("Starting server at Port: " + os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
