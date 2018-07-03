package main

import (
	"github.com/gin-gonic/gin"
	"citicab/app"
	"os"
	"melody"
	"encoding/json"
	"citicab/core"
	"citicab/controllers"
	"citicab/models"
)

var (
	DRIVER_SUB = "driver_sub"
	RIDE_SUB = "ride_sub"
	EVENT = "events"
)
func main() {

	r := gin.New()
	m := melody.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(app.GinJwt)
	gin.SetMode(gin.ReleaseMode)

	r.POST("/api/user/login", controllers.UserLogin)
	r.POST("/api/driver/login", controllers.DriverLogin)
	r.POST("/api/user/verify", controllers.VerifyUser)
	r.POST("/api/driver/verify", controllers.VerifyDriver)
	r.POST("/api/driver/update", controllers.UpdateDriver)
	r.POST("/api/driver/password/update", controllers.UpdateDriversPassword)
	r.POST("/api/user/update", controllers.UpdateUser)
	r.POST("/api/user/password/update", controllers.UpdatePassword)
	r.POST("/api/user/code/verify", controllers.VerifyCode)
	r.POST("/api/driver/code/verify", controllers.VerifyDriverCode)
	r.POST("/api/ride/new", controllers.NewRide)
	r.POST("/api/ride/status/:ride", controllers.UpdateStatus)
	r.GET("/api/user/verify/resend", controllers.ResendOtpCode)
	r.GET("/api/driver/verify/resend", controllers.ResendDriverOtpCode)
	r.POST("/api//driver/location/update", controllers.UpdateLocation)
	r.POST("/api/driver/status/update", controllers.UpdateDriverStatus)

	r.GET("/api/ws/connect", func(context *gin.Context) {
		m.HandleRequest(context.Writer, context.Request)
	})

	m.HandleConnect(func(session *melody.Session) {

	})

	m.HandleMessage(func(session *melody.Session, bytes []byte) {

		var data models.IncomingMessage
		err := json.Unmarshal(bytes, &data)
		if err == nil {

			action := data.Action
			switch action {
			case DRIVER_SUB:
				dv := models.GetDriver(data.UniqueId)
				if dv != nil {
					core.SubscribeDriverToChannel(dv, session)
				}
				break
			case RIDE_SUB:
				ride := models.GetRide(data.UniqueId)
				if ride != nil {
					core.SubscribeToRideChannel(ride, session)
				}
				break
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8009"
	}

	r.Run(":" + port)
}
