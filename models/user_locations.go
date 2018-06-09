package models

import u "citicab/utils"

type UserLocation struct {
	UserId uint `json:"user_id"`
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func UpdateLocation(loc *UserLocation) (map[string]interface{}) {

	err := Db.Table("user_locations").Where("user_id = ?", loc.UserId).UpdateColumn(loc).Error
	if err != nil {
		return u.Message(false, "Failed to update location. Please, retry")
	}

	return u.Message(true, "location updated")
}

