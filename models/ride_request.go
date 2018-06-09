package models

import "encoding/json"

type RideRequest struct {
	FromLat json.Number `json:"from_lat"`
	ToLat json.Number `json:"to_lat"`
	FromLon json.Number `json:"from_lon"`
	ToLon json.Number `json:"to_lon"`
	PickUpAddress string `json:"pick_up_address"`
	DestinationAddress string `json:"destination_address"`
	Message string `json:"message"`
}
