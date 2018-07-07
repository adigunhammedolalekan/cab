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
	fmt.Println(sessId)
	sess := wsChannels[sessId]
	if sess != nil {
		wsMessage := &models.WsMessage{
			Action: "new_ride",
			NewRide: ride,
		}

		data, _ := json.Marshal(wsMessage)
		sess.Write(data)
	}
}

func SubscribeDriverToChannel(driver *models.Driver, session *melody.Session) bool {

	sessId := fmt.Sprintf("driver%d", driver.ID)
	_, ok := wsChannels[sessId]
	if !ok {
		wsChannels[sessId] = session
		session.Write([]byte("Subscribed"))
		return true
	}

	fmt.Println(len(wsChannels))
	//Already subscribed
	return false
}


func SubscribeToRideChannel(ride *models.Ride, session *melody.Session) bool {

	sessId := fmt.Sprintf("ride%d", ride.ID)
	_, ok := rideChannels[sessId]
	if !ok {
		channel := &models.Channel{
			ChannelName: sessId,
		}
		channel.Sessions = append(channel.Sessions, session)
		rideChannels[sessId] = channel
		return true
	}

	//Already subscribed
	return false
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

	sessId := fmt.Sprintf("ride%d", ride.ID)
	channel, ok := rideChannels[sessId]
	if ok {
		return channel.Send(ride)
	}

	return nil
}