package models

import (
	"errors"
	"net/http"
	"squabble/db"
	"squabble/error_constant"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

type TokenDetails struct {
	SessionUUID string `gorm:"primaryKey"`
	Username    string
}

var authModel AuthModel

func GetAuthModel() AuthModel {
	return authModel
}

type AuthModel struct{}

func (m AuthModel) CreateToken(username string) (*TokenDetails, error) {

	td := &TokenDetails{}
	td.Username = username
	td.SessionUUID = uuid.New().String()

	if err := db.GetRedis().Set(td.SessionUUID, td.Username, time.Hour*24*7).Err(); err != nil {
		return nil, err
	}
	return td, nil
}

func (m AuthModel) VerifyToken(sessionUUID string) (string, error) {
	userid, err := db.GetRedis().Get(sessionUUID).Result()
	if err == redis.Nil {
		return "", errors.New(error_constant.VerifyTokenInvalid)
	}
	if err != nil {
		return "", errors.New(error_constant.VerifyTokenFailed)
	}
	return userid, nil
}

func (m AuthModel) DeleteToken(sessionUUID string) error {
	if _, err := db.GetRedis().Del(sessionUUID).Result(); err != nil {
		return err
	}
	return nil
}

func SessionAuth(w http.ResponseWriter, r *http.Request) (string, error) {
	sessionUUID := r.Header.Get("session-id")
	return GetUsername(sessionUUID)
}

func GetUsername(sessionUUID string) (string, error) {
	username, err := GetUserModel().VerifyToken(sessionUUID)
	if err != nil {
		return "", err
	}
	return username, nil
}
