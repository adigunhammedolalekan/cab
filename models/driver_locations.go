package models

import u "citicab/utils"
type DriverLocation struct {
	DriverId uint `json:"driver_id"`
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}


func UpdateDriversLocation(loc *DriverLocation) (map[string]interface{}) {

	err := Db.Table("user_locations").Where("user_id = ?", loc.DriverId).UpdateColumn(loc).Error
	if err != nil {
		return u.Message(false, "Failed to update location. Please, retry")
	}

	return u.Message(true, "location updated")
}

func GetDriversLocation(id uint) *DriverLocation {

	loc := &DriverLocation{}
	err := Db.Table("driver_locations").Where("driver_id = ?", id).First(loc).Error
	if err != nil {
		return nil
	}

	return loc
}