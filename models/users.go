package models

import (
	"github.com/jinzhu/gorm"
	"strings"
	u "citicab/utils"
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"errors"
)

type User struct {
	gorm.Model
	Fullname string `json:"fullname"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Password string `json:"password"`
	Status string `json:"status"`
	Picture string `json:"picture"`
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
		text := fmt.Sprintf("Your CitiKab authentication code: %d", auth.Code)
		smsRequest := &SmsRequest{
			To: phone,
			DND: "1",
			Body: text,
			From: "CitiKab",
			ApiToken: "evNlSXxvpzkJyzAVadcH024byBSqZbLiTAI80YbgRYzIaphAR4bUuWyTW63J",
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

	user = temp;
	user.Password = ""
	user.Token = GenJWT(user.ID)
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


func ChangeUsersPassword(old, newPassword string, id uint) error {

	if len(old) <= 0 || len(newPassword) <= 0 {
		return errors.New("Password body cannot be empty")
	}

	user := GetUser(id)
	if user == nil {
		return errors.New("User not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(old))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.New("Old password doesn't match")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	err = Db.Table("users").Where("id = ?", id).UpdateColumn("password", string(hashed)).Error
	if err != nil {
		return errors.New("Failed to update password at this time. Please retry")
	}

	return nil
}

func GetUserRideHistory(id uint) (error, []*Ride) {

	temp := make([]*Ride, 0)
	err := Db.Table("rides").Where("user_id = ?", id).Find(&temp).Error
	if err != nil {
		return err, nil
	}

	data := make([]*Ride, 0)
	for _, next := range temp {
		data = append(data, GetRide(next.ID))
	}

	return nil, data
}


func EditUser(column, value string, id uint) (error, *User){

	user := &User{}
	if column == "email" {
		Db.Table("users").Where("email = ?", value).First(user)
		if user.ID > 0 {
			return errors.New("Email already in use by another customer"), nil
		}
	}
	if column == "phone" {
		Db.Table("users").Where("phone = ?", value).First(user)
		if user.ID > 0 {
			return errors.New("Phone number already in use by another customer"), nil
		}
	}

	Db.Table("users").Where("id = ?", id).UpdateColumn(column, value)

	acc := GetUser(id)
	if acc != nil {
		acc.Password = ""
	}
	return nil, acc
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

type Card struct {
	gorm.Model
	UserId uint `json:"user_id"`
	CardNo string `json:"card_no"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear string `json:"expiry_year"`
	Cvv string `json:"cvv"`
}

func AddCard(card *Card) error {

	if card.UserId <= 0 {
		return errors.New("Card must have a user!")
	}

	return Db.Create(card).Error
}

func GetCards(user uint) ([]*Card, error) {

	data := make([]*Card, 0)
	err := Db.Table("cards").Where("user_id = ?", user).Find(&data).Error
	if err != nil {
		return nil, err
	}

	return data, nil
}

