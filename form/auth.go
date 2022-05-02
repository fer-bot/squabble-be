package form

import (
	"squabble/models"
)

type LoginFormRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginFormResponse struct {
	Username    string `json:"username"`
	SessionUUID string `json:"session-id"`
}

func LoginFormResponseBuilder(tokenDetails *models.TokenDetails) *LoginFormResponse {
	r := LoginFormResponse{}
	r.Username = tokenDetails.Username
	r.SessionUUID = tokenDetails.SessionUUID
	return &r
}

type RegisterFormRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UsernameFormResponse struct {
	Username string `json:"username"`
}

func UsernameFormResponseBuilder(username string) *UsernameFormResponse {
	r := UsernameFormResponse{}
	r.Username = username
	return &r
}
