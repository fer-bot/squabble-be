package controller

import (
	"encoding/json"
	"net/http"
	"squabble/form"
	"squabble/models"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginFormRequest form.LoginFormRequest
	err := json.NewDecoder(r.Body).Decode(&loginFormRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}
	tokenDetails, err := models.GetUserModel().Login(loginFormRequest.Username, loginFormRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}
	json.NewEncoder(w).Encode(form.LoginFormResponseBuilder(tokenDetails))
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, username string) {
	sessionUUID := r.Header.Get("session-id")
	if err := models.GetUserModel().Logout(sessionUUID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}
	json.NewEncoder(w).Encode(form.UsernameFormResponseBuilder(username))
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var registerFormRequest form.RegisterFormRequest
	err := json.NewDecoder(r.Body).Decode(&registerFormRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}
	user, err := models.GetUserModel().Register(registerFormRequest.Username, registerFormRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(form.SingleErrorResponseBuilder(err))
		return
	}
	json.NewEncoder(w).Encode(form.UsernameFormResponseBuilder(user.Username))
}
