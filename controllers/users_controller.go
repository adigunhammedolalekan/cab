package controllers

import (
	"github.com/gin-gonic/gin"
	u "citicab/utils"
	"citicab/models"
	"fmt"
	"strconv"
)

var VerifyUser = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	phone := data["phone"]
	r := models.VerifyPhone(phone . (string))
	c.JSON(200, r)
}

var VerifyCode = func(c *gin.Context) {

	authCode := &models.AuthCode{}
	err := c.ShouldBind(authCode)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	code, err := authCode.Code.Int64()
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	user, ok := c.Get("user")
	id := user . (uint)
	if !ok || id <= 0 {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	auth := models.GetAuth(id)
	if auth == nil {
		c.JSON(200, u.Message(false, "No authentication code for user"))
		return
	}
	if auth.Code != int(code) {
		c.JSON(200, u.Message(false, "Code does not match. Please retry"))
		return
	}

	c.JSON(200, u.Message(true, "Success."))
}

var UpdateUser = func(c *gin.Context) {

	user := &models.User{}
	err := c.ShouldBind(user)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	user.ID = id . (uint)
	user.Password = ""
	r := user.Update()
	c.JSON(200, r)
}

var UpdatePassword = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	password := data["password"] . (string)
	user := &models.User{}
	user.ID = id . (uint)
	user.Password = password
	r := models.UpdatePassword(user)
	c.JSON(200, r)
}

var UserLogin = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	phone := data["phone"] . (string)
	user := &models.User{Phone: phone, Password: data["password"] . (string)}
	r := models.Login(user)
	c.JSON(200, r)
}

var ResendOtpCode = func(c *gin.Context) {

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	user := id . (uint)
	auth := models.CreateAuth(user)
	if auth != nil {
		acc := models.GetUser(user)
		if acc != nil {
			auth.SendToUser(acc.Phone)
		}
	}

	c.JSON(200, u.Message(true, "Code Sent!"))
}

var UserForgotPassword = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	email := data["email"]
	user := models.GetUserByEmail(email . (string))

	r := u.Message(false, fmt.Sprintf("User with email %s not found", email . (string)))
	if user != nil {
		r = user.SendForgotPasswordEmail()
	}
	c.JSON(200, r)
}


var ChangeUserPassword = func(c *gin.Context) {

	req := &models.ChangePasswordRequest{}
	err := c.ShouldBind(req)
	if err != nil {
		c.AbortWithStatusJSON(403, u.InvalidRequestMessage())
		return
	}
	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	err = models.ChangeUsersPassword(req.OldPassword, req.NewPassword, id)
	if err != nil {
		c.JSON(200, u.Message(false, err.Error()))
		return
	}

	c.JSON(200, u.Message(true, "success"))
}


var GetUserRideHistory = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	err, data := models.GetUserRideHistory(id)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, "Failed to fetch history. Please retry"))
		return
	}

	response := u.Message(true, "success")
	response["data"] = data
	c.JSON(200, response)
}


var EditUserAccount = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	column := data["column"] . (string)
	value := data["value"] . (string)
	err, person := models.EditUser(column, value, id)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	response := u.Message(true, "success")
	response["data"] = person
	c.JSON(200, response)
}

var AddCard = func(c *gin.Context) {

	card := &models.Card{}
	err := c.ShouldBind(card)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	card.UserId = id
	err = models.AddCard(card)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	c.JSON(200, u.Message(true, "success"))
}

var GetCards = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	data, err := models.GetCards(id)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, "error"))
		return
	}

	response := u.Message(true, "success")
	response["data"] = data
	c.JSON(200, response)
}

var RemoveCard = func(c *gin.Context) {

	_, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	card, ok := c.Params.Get("card")
	if !ok {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	cid, err := strconv.Atoi(card)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	cardId := uint(cid)
	models.RemoveCard(cardId)
	c.JSON(200, u.Message(true, "success"))
}
