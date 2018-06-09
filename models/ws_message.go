package models

type WsMessage struct {

	Action string `json:"action"`
	NewRide *Ride `json:"new_ride"`
	ChannelName string `json:"channel_name"`

}

type IncomingMessage struct {

	Action string `json:"action"`
	UniqueId uint `json:"unique_id"`

}