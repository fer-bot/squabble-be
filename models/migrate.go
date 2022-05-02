package models

import (
	"squabble/db"
)

func AutoMigrate() {
	db.GetDB().AutoMigrate(&User{})
}
