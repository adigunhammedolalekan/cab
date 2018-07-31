package models

import (
	"github.com/jinzhu/gorm"
	"errors"
	"github.com/rpip/paystack-go"
	"os"
	"encoding/json"
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

func GetBanks() (*paystack.BankList, error) {

	ps := paystack.NewClient(os.Getenv("PS_KEY"), nil)
	return ps.Bank.List()
}

type TransferRequest struct {

	DriverId uint `json:"driver_id"`
	AccountNumber string `json:"account_number"`
	BankCode string `json:"bank_code"`
	BankName string `json:"bank_name"`
	Name string `json:"name"`
	Amount json.Number `json:"amount"`
}

func (tr *TransferRequest) ValidAmount() (float64) {

	data, err := tr.Amount.Float64()
	if err != nil {
		return 0.0;
	}

	return data
}

func CreateTransfer(request *TransferRequest) (error) {

	amt := request.ValidAmount()
	w := GetWallet(request.DriverId)

	if float32(amt) > w.Balance {
		return errors.New("Failed to perform transfer. Insufficient funds")
	}

	rc := &paystack.TransferRecipient{}
	rc.AccountNumber = request.AccountNumber
	rc.Name = request.Name
	rc.BankCode = request.BankCode
	rc.Type = "nuban"

	ps := paystack.NewClient(os.Getenv("PS_KEY"), nil)
	response, err := ps.Transfer.CreateRecipient(rc)
	if err != nil {
		return err
	}

	treq := &paystack.TransferRequest{}
	treq.Amount = float32(request.ValidAmount() / 100)
	treq.Recipient = response.RecipientCode
	treq.Source = "balance"

	transfer, err := ps.Transfer.Initiate(treq)
	if err != nil {
		return err
	}

	if transfer.Status == "success" {
		tf := &Transfer{}
		tf.Amount = float64(treq.Amount)
		tf.DriverId = request.DriverId
		tf.AccountNumber = request.AccountNumber
		tf.BankName = request.BankName
		tf.BankCode = request.BankCode

		newBal := w.Balance - float32(amt)
		tx := Db.Begin()
		err := tx.Table("wallets").Where("driver_id = ?", request.DriverId).UpdateColumn("balance", newBal).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Create(tf).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()
	}

	return errors.New("Failed to complete transfer request. Please retry")
}

type Transfer struct {

	gorm.Model
	DriverId uint `json:"driver_id"`
	Amount float64 `json:"amount"`
	AccountNumber string `json:"account_number"`
	BankCode string `json:"bank_code"`
	BankName string `json:"bank_name"`

}