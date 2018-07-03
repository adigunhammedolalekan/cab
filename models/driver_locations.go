package models

import (
	u "citicab/utils"
	"github.com/jinzhu/gorm"
)
type DriverLocation struct {
	gorm.Model
	DriverId uint `json:"driver_id"`
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}


func UpdateDriversLocation(loc *DriverLocation) (map[string]interface{}) {

	temp := &DriverLocation{}
	err := Db.Table("driver_locations").Where("driver_id = ?", loc.DriverId).First(temp).Error
	if err == gorm.ErrRecordNotFound {
		err = Db.Create(loc).Error
		return u.Message(err == nil, "location updated")
	}

	err = Db.Table("driver_locations").Where("driver_id = ?", loc.DriverId).UpdateColumn(loc).Error
	if err != nil {
		return u.Message(false, "Failed to update location. Please, retry")
	}

	return u.Message(true, "location updated")
}

func GetDriversLocation(id uint) *DriverLocation {

	loc := &DriverLocation{}
	err := Db.Table("driver_location").Where("driver_id = ?", id).First(loc).Error
	if err != nil {
		return nil
	}

	return loc
}