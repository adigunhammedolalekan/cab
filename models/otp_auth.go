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

	var count int
	err := Db.Table("auths").Where("user_id = ?", user).Count(&count).Error

	if count == 0 {
		err = Db.Create(auth).Error
		if err == nil {
			fmt.Println(auth.Code)
			return auth
		}

		return nil
	}else {
		err = Db.Table("auths").Where("user_id = ?", auth.UserId).UpdateColumn(auth).Error
		if err != nil {
			return nil
		}
	}
	fmt.Println(auth.Code)
	return auth
}

func (auth *Auth) SendToUser(phone string) {

	text := fmt.Sprintf("Your CitiCab authentication code is %d", auth.Code)
	smsRequest := &SmsRequest{
		Text: text,
		Phone: phone,
	}

	SmsQueue <- smsRequest
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
