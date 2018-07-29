package models

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Wallet struct {
	gorm.Model
	DriverId uint `json:"driver_id"`
	Balance float32 `json:"balance"`
}

func NewWallet(driver uint) *Wallet {
	w := &Wallet{}
	w.DriverId = driver
	w.Balance = 0.0

	return w
}

func FundWallet(driver uint, amount float32) error {

	wallet := GetWallet(driver)
	if wallet == nil {
		return errors.New("Wallet not found for driver")
	}

	wallet.Balance += amount
	return Db.Table("wallets").Where("driver_id = ?", driver).Update(wallet).Error
}

func GetWallet(dv uint) *Wallet {

	w := &Wallet{}
	err := Db.Table("wallets").Where("driver_id = ?", dv).First(w).Error
	if err != nil {
		return nil
	}

	return w
}
