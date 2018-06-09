package models

import (
	"github.com/jinzhu/gorm"
	"math/rand"
	"fmt"
)

type Auth struct {
	gorm.Model
	UserId uint `json:"user_id"`
	Code int `json:"code"`
}

func CreateAuth(user uint) *Auth {

	auth := &Auth{
		UserId: user,
		Code: rand.Intn(99999),
	}

	err := Db.FirstOrCreate(auth, "user = ?", user).Error
	if err != nil {
		return nil
	}

	return auth
}


func GetAuth(user uint) *Auth {

	auth := &Auth{}
	err := Db.Table("auths").Where("user_id = ?", user).First(auth).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return auth
}
