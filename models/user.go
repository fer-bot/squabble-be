package models

import (
	"errors"
	"time"

	"squabble/db"
	"squabble/error_constant"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username  string    `gorm:"primaryKey"`
	Password  []byte    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

var userModel UserModel

func GetUserModel() *UserModel {
	return &userModel
}

type UserModel struct{}

func (m UserModel) Login(username string, password string) (*TokenDetails, error) {
	var user User
	result := db.GetDB().Where("username = ?", username).Find(&user)
	if result.Error != nil {
		return &TokenDetails{}, errors.New(error_constant.LoginFailed)
	}
	if result.RowsAffected == 0 {
		return &TokenDetails{}, errors.New(error_constant.LoginInvalid)
	}

	bytePassword := []byte(password)
	if err := bcrypt.CompareHashAndPassword(user.Password[:], bytePassword); err != nil {
		return &TokenDetails{}, errors.New(error_constant.LoginInvalid)
	}
	return authModel.CreateToken(username)
}

func (m UserModel) Register(username string, password string) (*User, error) {
	var user User
	result := db.GetDB().Where("username = ?", username).Find(&user)
	if result.Error != nil {
		return &User{}, errors.New(error_constant.RegisterFailed)
	}
	if result.RowsAffected > 0 {
		return &User{}, errors.New(error_constant.RegisterInvalid)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &User{}, errors.New(error_constant.RegisterFailed)
	}
	user.Username = username
	user.Password = hashedPassword

	result = db.GetDB().Create(&user)
	if result.Error != nil {
		return &User{}, errors.New(error_constant.RegisterFailed)
	}

	return &user, nil
}

func (m UserModel) Logout(sessionUUID string) error {
	if err := authModel.DeleteToken(sessionUUID); err != nil {
		return err
	}
	return nil
}

func (m UserModel) VerifyToken(sessionUUID string) (string, error) {
	userid, err := authModel.VerifyToken(sessionUUID)
	if err != nil {
		return "", err
	}
	return userid, nil
}
