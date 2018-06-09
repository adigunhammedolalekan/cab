package models

import (
	"github.com/jinzhu/gorm"
	"github.com/kellydunn/golang-geo"
	"fmt"
	u "citicab/utils"
)

type Ride struct {

	gorm.Model
	UserId uint `json:"user_id"`
	DriverId uint `json:"driver_id"`
	Fee uint `json:"fee"`
	PickUpLat float64 `json:"pick_up_lat"`
	PickUpLon float64 `json:"pick_up_lon"`
	DestinationLat float64 `json:"destination_lat"`
	DestinationLon float64 `json:"destination_lon"`
	PickupAddress string `json:"pickup_address"`
	DestinationAddress string `json:"destination_address"`
	Status uint `json:"status"`
	Message string `json:"message"`
	User *User `gorm:"-" sql:"-" json:"user"`
	Driver *Driver `gorm:"-" sql:"-" json:"driver"`

}

func (ride *Ride) GetStatus() string {

	switch ride.Status {
	case 1:
		return "started"
	case 2:
		return "accepted"
	case 3:
		return "cancelled"
	case 4:
		return "ended"
	}

	return ""
}

func (ride *Ride) UpdateStatus() (error) {

	err := Db.Table("rides").Where("id = ?", ride.ID).UpdateColumn("status", ride.Status).Error
	if err != nil {
		return err
	}

	return nil
}

func CreateRide(ride *Ride) bool {

	user := GetUser(ride.UserId)
	driver := GetDriver(ride.DriverId)
	if user == nil || driver == nil {
		return false
	}

	ride.Status = 0
	err := Db.Create(ride).Error
	return err == nil
}

func FindDriver(loc *UserLocation) *Driver {

	mapper, err := geo.HandleWithSQL()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	point := geo.NewPoint(loc.Lat, loc.Lon)
	rows, err := mapper.PointsWithinRadius(point, float64(30))
	if err != nil {
		return nil
	}

	drivers := []*Driver{}
	u.MapRowsToSliceOfStruct(rows, &drivers, false)

	found := &Driver{}
	for _, d := range drivers {
		if d.Status == "online" {
			found = d
			break
		}
	}

	return found
}

func GetRide(id uint) *Ride {

	ride := &Ride{}
	err := Db.Table("rides").Where("id = ?", id).First(ride).Error
	if err != nil {
		return nil
	}

	ride.User = GetUser(ride.UserId)
	ride.Driver = GetDriver(ride.DriverId)
	return ride
}
