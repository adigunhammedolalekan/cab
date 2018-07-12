package controllers

import (
	"github.com/gin-gonic/gin"
	"citicab/models"
	u "citicab/utils"
	"citicab/core"
	"strconv"
)

var NewRide = func(c *gin.Context) {

	rideRequest := &models.RideRequest{}
	err := c.ShouldBind(rideRequest)
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	loc := &models.UserLocation{}
	lat, err := rideRequest.FromLat.Float64()
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}
	lon, err := rideRequest.FromLon.Float64()
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	toLat, err := rideRequest.ToLat.Float64()
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	toLon, err := rideRequest.ToLat.Float64()
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	loc.Lon = lon
	loc.Lat = lat
	user := id. (uint)
	ride := &models.Ride{
		UserId: user,
		PickUpLat: lat,
		PickUpLon: lon,
		DestinationLat: toLat,
		DestinationLon: toLon,
		DestinationAddress: rideRequest.DestinationAddress,
		PickupAddress: rideRequest.PickUpAddress,
		Message: rideRequest.Message,
	}

	driver := models.FindDriver(loc)
	if driver == nil || driver.ID <= 0 {
		c.JSON(200, u.Message(false, "No Driver found within your area"))
		return
	}

	ride.DriverId = driver.ID
	ok = models.CreateRide(ride)
	if !ok {
		c.JSON(200, u.Message(false, "Failed to create ride. Please retry"))
		return
	}

	ride = models.GetRide(ride.ID)
	r := u.Message(true, "Ride Created")
	r["ride"] = ride

	core.NotifyDriver(ride)
	c.JSON(200, r)
}

var UpdateStatus = func(c *gin.Context) {

	id, ok := c.Get("user")
	if !ok {
		c.JSON(200, u.UnAuthorizedMessage())
		return
	}

	user := id . (uint)
	rideId, err := strconv.Atoi(c.Param("ride"))
	if err != nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	ride := models.GetRide(uint(rideId))
	if ride == nil {
		c.JSON(200, u.InvalidRequestMessage())
		return
	}

	if ride.UserId == user || ride.DriverId == user {
		rs, err := strconv.Atoi(c.Param("status"))
		if err != nil {
			c.JSON(200, u.InvalidRequestMessage())
			return
		}

		ride.Status = uint(rs)
		err = ride.UpdateStatus()

		core.NotifyRideStatus(ride)
		c.JSON(200, u.Message(true, "Ride Status Updated"))
	}
}