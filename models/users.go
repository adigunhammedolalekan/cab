package models

import (
	"github.com/jinzhu/gorm"
	"strings"
	u "citicab/utils"
	"golang.org/x/crypto/bcrypt"
	"fmt"
)

type User struct {
	gorm.Model
	Fullname string `json:"fullname"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Password string `json:"password"`
	Status string `json:"status"`
	Verified bool `json:"verified"`
	Token string `gorm:"-" sql:"-" json:"token"`
}


func VerifyPhone(phone string) (map[string] interface{}) {

	var count int
	err := Db.Table("users").Where("phone = ?", strings.TrimSpace(phone)).Count(&count).Error
	if err != nil {
		return u.Message(false, "Failed to perform operation. Please, retry")
	}
	if count > 0 {
		resp := u.Message(true, "success")
		resp["exists"] = true
		return resp
	}

	user := &User{Phone: phone}
	err = Db.Create(user).Error
	if err != nil {
		return u.Message(false, "Failed to create new account. Please, retry")
	}

	auth := CreateAuth(user.ID)
	if auth != nil {
		text := fmt.Sprintf("Your CitiCab authentication code: %d", auth.Code)
		smsRequest := &SmsRequest{
			Text: text,
			Phone: strings.TrimSpace(phone),
		}

		SmsQueue <- smsRequest
	}
	token := GenJWT(user.ID)
	resp := u.Message(true, "success")
	resp["exists"] = false
	resp["token"] = token
	return resp
}

func (user *User) Update() (map[string] interface{}) {

	if len(user.Password) > 0 {
		return u.Message(false, "Cannot update password")
	}

	err := Db.Table("users").Where("id = ?", user.ID).UpdateColumn(user).Error
	if err != nil {
		return u.Message(false, "Failed to update account. Please, retry")
	}

	mailRequest := &MailRequest{
		Subject: "Welcome to CitiCab",
		Body: "Hi, Welcome to our new shining platform, citicab",
		To: user.Email,
	}

	MailQueue <- mailRequest

	updated := GetUser(user.ID)
	r := u.Message(true, "account updated")

	updated.Password = ""
	r["user"] = updated
	return r
}

func Login(user *User) (map[string]interface{}) {

	temp := &User{}
	err := Db.Table("users").Where("phone = ?", user.Phone).First(temp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "User with phone " + user.Phone + " not found")
		}
		return u.Message(false, "Failed to complete login request. Please, retry")
	}

	if temp.ID <= 0 {
		return u.Message(false, "User with phone " + user.Phone + " not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(temp.Password), []byte(user.Password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(false, "Invalid login credentials")
	}

	user.Password = ""
	user.Token = GenJWT(temp.ID)
	r := u.Message(true, "success")
	r["user"] = user
	return r
}

func UpdatePassword(user *User) (map[string]interface{}) {

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	err := Db.Table("users").Where("id = ?", user.ID).UpdateColumn(user).Error
	if err != nil {
		return u.Message(false, "Failed to update password. Please, retry")
	}

	return u.Message(true, "Password updated")
}

func (user *User) SendForgotPasswordEmail() (map[string] interface{}) {

	if user.ID <= 0 {
		return u.Message(false, "Set user first")
	}

	auth := CreateAuth(user.ID)
	if auth != nil {

		mailReq := &MailRequest{
			Subject: "Password Reset Instruction",
			Body: fmt.Sprintf("Hi, Use code %d to reset your password", auth.UserId),
			To: user.Email,
		}

		MailQueue <- mailReq
	}

	return u.Message(true, "Success")
}

func GetUserByEmail(email string) *User {

	user := &User{}
	err := Db.Table("users").Where("email = ?", strings.TrimSpace(email)).First(user).Error
	if err != nil {
		return nil
	}

	return user
}

func GetUser(id uint) *User {

	fmt.Println(id)
	user := &User{}
	err := Db.Table("users").Where("id = ?", id).First(user).Error
	if err != nil {
		return nil
	}

	return user
}
