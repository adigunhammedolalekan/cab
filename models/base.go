package models

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jinzhu/gorm"
	"os"
	"github.com/joho/godotenv"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"math/rand"
	"time"
	"encoding/json"
	"gopkg.in/njern/gonexmo.v2"
	"github.com/dgrijalva/jwt-go"
)

var (
	Db *gorm.DB
	SmsQueue = make(chan *SmsRequest, 1000)
	MailQueue = make(chan *MailRequest, 1000)
)

func init() {

	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")


	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)
	fmt.Println(dbUri)

	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		fmt.Print(err)
	}

	rand.Seed(time.Now().UnixNano())

	Db = conn
	Db.Debug().AutoMigrate(&User{}, &Driver{}, &DriverLocation{}, &UserLocation{}, &Auth{}, &Ride{})

	go MessageWorker()
}

func MessageWorker() {

	for {
		select {
		case m, ok := <- SmsQueue:
			if ok && m != nil {
				m.Send()
			}
			break
		case req, ok := <- MailQueue:
			if ok && req != nil {
				req.Send()
			}
			break
		}
	}
}

func GetDB() *gorm.DB {
	return Db
}

type AuthCode struct {
	Code json.Number `json:"code"`
}

type MailRequest struct {

	Subject string `json:"subject"`
	Body string `json:"body"`
	To string `json:"to"`
	Name string `json:"name"`

}

type SmsRequest struct {
	Text string `json:"text"`
	Phone string `json:"phone"`
}

func (smsRequest *SmsRequest) Send() (error) {

	nex, err := nexmo.NewClient(os.Getenv("NEXMO_API_KEY"), os.Getenv("NEXMO_SECRET_KEY"))
	if err != nil || nex == nil {
		fmt.Println(err)
	}
	message := &nexmo.SMSMessage{

		From: "CitiCab",
		To: smsRequest.Phone,
		Text: smsRequest.Text,
		Class: nexmo.Standard,
	}

	resp, err := nex.SMS.Send(message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(resp.MessageCount)
	return nil
}

func (request *MailRequest) Send() (error) {
	return SendEmail(request)
}

func SendEmail(request *MailRequest) error {

	from := mail.NewEmail("CitiCab", os.Getenv("email"))
	to := mail.NewEmail(request.Name, request.To)

	message := mail.NewSingleEmail(from, request.Subject, to, request.Body, request.Body)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GenJWT(user uint) string {
	tk := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &Token{UserId: user})
	token, _ := tk.SignedString([]byte(os.Getenv("tk_password")))
	return token
}

