package models

import (
	"github.com/jinzhu/gorm"
	"github.com/kellydunn/golang-geo"
	"fmt"
	"errors"
)

const (
	EARTH_RADIUS = 6371
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
	Status uint `json:"status"` //0 = new, 1 = rejected, 2 = driver_cancelled, 3 = user_cancelled, 4 = started, 5 = ended
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

	tx := Db.Begin()

	err := tx.Table("rides").Where("id = ?", ride.ID).UpdateColumn("status", ride.Status).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if ride.Status == 2 || ride.Status == 3 {//Cancelled by USER or DRIVER?, free the driver
		err = tx.Table("drivers").Where("id = ?", ride.Driver.ID).UpdateColumn("occupied", 0).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func CreateRide(ride *Ride) bool {

	user := GetUser(ride.UserId)
	driver := GetDriver(ride.DriverId)
	if user == nil || driver == nil {
		fmt.Println("Driver or User is nil")
		return false
	}

	ride.Status = 0
	err := Db.Create(ride).Error

	fmt.Println(err)
	return err == nil
}

func FindDriver(loc *UserLocation) *Driver {

	sql := BuildSQL(loc.Lat,loc.Lon, float64(30.00))
	driverLocations := make([]*DriverLocation, 0)
	err := Db.Raw(sql).Scan(&driverLocations).Error
	if err != nil {
		fmt.Println(err)
		return nil
	}

	nearestDriver := &Driver{}
	shortestDistance := 100.0
	for _, driverLoc := range driverLocations {

		next := GetDriver(driverLoc.DriverId)
		if next.Occupied == 1 || next.Status == "offline" { //Driver is offline or already occupied
			continue
		}

		p1 := geo.NewPoint(loc.Lat, loc.Lon)
		p2 := geo.NewPoint(driverLoc.Lat, driverLoc.Lon)
		distance := p1.GreatCircleDistance(p2)
		nearestDriver = next
		if distance < shortestDistance {
			shortestDistance = distance
			nearestDriver = next
		}
	}

	if nearestDriver.ID > 0 {
		nearestDriver.SetOccupied()
	}
	return nearestDriver
}

func GetDriverRideHistory(id uint) (error, []*Ride) {

	temp := make([]*Ride, 0)
	err := Db.Table("rides").Where("driver_id = ?", id).Find(&temp).Error
	if err != nil {
		return err, nil
	}

	data := make([]*Ride, 0)
	for _, next := range temp {
		data = append(data, GetRide(next.ID))
	}

	return nil, data
}

func BuildSQL(lat, lon, radius float64) string {

	select_str := fmt.Sprintf("SELECT * FROM driver_locations a")
	lat1 := fmt.Sprintf("sin(radians(%f)) * sin(radians(a.lat))", lat)
	lng1 := fmt.Sprintf("cos(radians(%f)) * cos(radians(a.lat)) * cos(radians(a.lon) - radians(%f))", lat, lon)
	where_str := fmt.Sprintf("WHERE acos(%s + %s) * %f <= %f", lat1, lng1, float64(EARTH_RADIUS), radius)
	query := fmt.Sprintf("%s %s", select_str, where_str)

	return query
}

func GetRide(id uint) *Ride {

	ride := &Ride{}
	err := Db.Table("rides").Where("id = ?", id).First(ride).Error
	if err != nil {
		return nil
	}

	user := GetUser(ride.UserId)
	driver := GetDriver(ride.DriverId)

	if user != nil {
		user.Password = ""
	}

	if driver != nil {
		driver.Password = ""
	}

	ride.User = user
	ride.Driver = driver

	return ride
}

type Rating struct {
	gorm.Model
	RideId uint `json:"ride_id"`
	Rating float64 `json:"rating"`
	Comment string `json:"comment"`
	DriverId uint `json:"driver_id"`
	UserId uint `json:"user_id"`

	Ride *Ride `sql:"-" gorm:"-" json:"ride"`
}

func (r *Rating) Create() (error) {

	var count int
	err := Db.Table("ratings").Where("ride_id = ?", r.RideId).Count(&count).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return Db.Create(r).Error
	}

	if count > 0 {
		return errors.New("Ride has already been rated")
	}

	return nil
}

func GetRating(driver uint) []*Rating {

	data := make([]*Rating, 0)
	err := Db.Table("ratings").Where("driver_id = ?", driver).Find(&data).Error
	if err != nil {
		return nil
	}

	result := make([]*Rating, 0)
	for _, v := range data {
		if v != nil {
			v.Ride = GetRide(v.RideId)
			result = append(result, v)
		}
	}

	return result
}
