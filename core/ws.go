package core

import (
	"citicab/models"
	"melody"
	"fmt"
	"encoding/json"
)

var (
	wsChannels = make(map[string] *melody.Session)
	rideChannels = make(map[string] *models.Channel)
)
func NotifyDriver(ride *models.Ride) {

	sessId := fmt.Sprintf("driver%d", ride.DriverId)
	sess := wsChannels[sessId]
	if sess != nil {
		wsMessage := &models.WsMessage{
			Action: "new_ride",
			NewRide: ride,
		}

		data, _ := json.Marshal(wsMessage)
		err := sess.Write(data)
		fmt.Println(err)
	}
}

func SubscribeDriverToChannel(driver *models.Driver, session *melody.Session) {

	sessId := fmt.Sprintf("driver%d", driver.ID)
	wsChannels[sessId] = session

}


func SubscribeUserToChannel(user *models.User, session *melody.Session) {

	sessId := fmt.Sprintf("user%d", user.ID)
	wsChannels[sessId] = session

}

func UnSubscribeDriverFromChannel(driver *models.Driver) bool {

	sessId := fmt.Sprintf("driver%d", driver.ID)
	sess, ok := wsChannels[sessId]
	if ok && sess != nil {
		delete(wsChannels, sessId)
		return true
	}

	//Already unsubscribed
	return false
}

func UnSubscribeFromRideChannel(ride *models.Ride) bool {

	sessId := fmt.Sprintf("ride%d", ride.ID)
	ch, ok := rideChannels[sessId]
	if ok && ch != nil {
		delete(rideChannels, sessId)
		return true
	}

	//Already unsubscribed
	return false
}

func NotifyRideStatus(ride *models.Ride) error {


	wsMessage := &models.WsMessage{
		Action: "ride_update",
		NewRide: ride,
	}

	data, _ := json.Marshal(wsMessage)
	driverSessId := fmt.Sprintf("driver%d", ride.Driver.ID)
	userSessId := fmt.Sprintf("user%d", ride.User.ID)

	sess := wsChannels[driverSessId]
	if sess != nil {
		fmt.Println("Sending to driver")
		err := sess.Write(data)
		fmt.Println(err)
	}

	s := wsChannels[userSessId]
	if s != nil {
		fmt.Println("Sending to User")
		err := s.Write(data)
		fmt.Println(err)
	}

	return nil
}