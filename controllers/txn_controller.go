package controllers

import (
	"github.com/gin-gonic/gin"
	u "citicab/utils"
	models "citicab/models"
	"github.com/rpip/paystack-go"
	"os"
)

var InitTxn = func(c *gin.Context) {

	id, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	payload := &models.TxnRequestPayload{}
	err := c.ShouldBind(payload)
	if err != nil {
		c.AbortWithStatusJSON(403, u.InvalidRequestMessage())
		return
	}

	user := models.GetUser(id. (uint))
	if user == nil {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	if payload.AmountValue() <= 0 {
		c.AbortWithStatusJSON(400, u.Message(false, "Invalid Txn amount"))
		return
	}

	txnRequest := &paystack.TransactionRequest{}
	txnRequest.Amount = float32(payload.AmountValue() * 100 /*Convert amount to KOBO*/)
	txnRequest.Email = user.Email

	ps := paystack.NewClient(os.Getenv("PS_KEY"), nil)
	response, err := ps.Transaction.Initialize(txnRequest)
	if err != nil {
		c.JSON(200, u.Message(false, err.Error()))
		return
	}

	resp := u.Message(true, "success")
	resp["access_code"] = response["access_code"] . (string)
	c.JSON(200, resp)
}

var VerifyTxn = func(c *gin.Context) {

	id, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	user := models.GetUser(id. (uint))
	if user == nil {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	ps := paystack.NewClient(os.Getenv("PS_KEY"), nil)
	txn, err := ps.Transaction.Verify(data["ref"] . (string))
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, "Unable to verify transaction."))
		return
	}

	rId := data["ride_id"] . (float64)
	ride := models.GetRide(uint(rId))
	if ride == nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	if txn.Status == "success" {
		value := txn.Amount / 100.0
		err := models.FundWallet(ride.DriverId, value)
		if err != nil {
			c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
			return
		}
		c.JSON(200, u.Message(true, "success"))
	}else {
		c.JSON(200, u.Message(false, "Invalid transaction"))
	}
}

var GetBanks = func(c *gin.Context) {

	banks, err := models.GetBanks()
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, "Failed to fetch bank data"))
		return
	}

	data := u.Message(true, "success")
	data["data"] = banks.Values
	c.JSON(200, data)
}