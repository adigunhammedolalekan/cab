package controllers

import (
	"github.com/gin-gonic/gin"
	"citicab/models"
	u "citicab/utils"
)

var VerifyDriver = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	phone := data["phone"]
	r := models.VerifyDriversPhone(phone . (string))
	c.JSON(200, r)
}


var VerifyDriverCode = func(c *gin.Context) {

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
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	auth := models.GetAuth(id)
	if auth == nil {
		c.JSON(200, u.Message(false, "No authentication code user"))
		return
	}
	if auth.Code != int(code) {
		c.JSON(200, u.Message(false, "Code does not match. Please retry"))
		return
	}

	c.JSON(200, u.Message(true, "Success."))
}

var UpdateDriver = func(c *gin.Context) {

	driver := &models.Driver{}
	err := c.ShouldBind(driver)
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

	driver.ID = id
	driver.Password = ""
	r := driver.Update()
	c.JSON(200, r)
}

var UpdateDriversPassword = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	password := data["password"] . (string)
	driver := &models.Driver{}
	user, ok := c.Get("user")
	id := user . (uint)
	if !ok || id <= 0 {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	driver.ID = id
	driver.Password = password
	r := models.UpdateDriversPassword(driver)
	c.JSON(200, r)
}

var DriverLogin = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	phone := data["phone"] . (string)
	driver := &models.Driver{Phone: phone, Password: data["password"] . (string)}
	r := models.DriverLogin(driver)
	c.JSON(200, r)
}


var ResendDriverOtpCode = func(c *gin.Context) {

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	user := id . (uint)
	auth := models.CreateAuth(user)
	if auth != nil {
		acc := models.GetDriver(user)
		if acc != nil {
			auth.SendToUser(acc.Phone)
		}
	}

	c.JSON(200, u.Message(true, "Code Sent!"))
}

var UpdateLocation = func(c *gin.Context) {

	driverLoc := &models.DriverLocation{}
	err := c.ShouldBind(driverLoc)
	if err != nil {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	user := id . (uint)
	driverLoc.DriverId = user
	m := models.UpdateDriversLocation(driverLoc)
	c.JSON(200, m)
}

var UpdateDriverStatus = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.UnAuthorizedMessage())
		return
	}

	user := id . (uint)
	err = models.UpdateDriverStatus(user, data["status"] . (string))

	var message string = "Failed to update status."
	if err == nil {
		message = "status updated"
	}
	c.JSON(200, u.Message(true, message))
}

var EditAccount = func(c *gin.Context) {

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	user, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	err, dv := models.Edit(data["column"] . (string), data["value"] . (string), id)
	if err != nil {
		c.JSON(200, u.Message(false, err.Error()))
		return
	}

	response := u.Message(true, "success")
	response["data"] = dv
	c.JSON(200, response)
}

var GetRideHistory = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.JSON(403, u.UnAuthorizedMessage())
		return
	}

	id := user . (uint)
	err, data := models.GetDriverRideHistory(id)
	if err != nil {
		c.JSON(200, u.Message(false, "Failed to fetch history. Please retry"))
		return
	}

	response := u.Message(true, "success")
	response["data"] = data
	c.JSON(200, response)
}
