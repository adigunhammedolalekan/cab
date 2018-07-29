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
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"net/url"
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
	Db.Debug().AutoMigrate(&User{}, &Driver{},
	&DriverLocation{}, &UserLocation{},
	&Auth{}, &Ride{}, &Card{}, &Rating{}, &Wallet{})

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

type TxnRequestPayload struct {
	Amount json.Number `json:"amount"`
}

func (p *TxnRequestPayload) AmountValue() (float64) {

	amt, err := p.Amount.Float64()
	if err != nil {
		return 0.0
	}

	return amt
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

	ApiToken string `json:"api_token"`
	To string `json:"to"`
	From string `json:"from"`
	Body string `json:"body"`
	DND string `json:"dnd"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	Driver uint `json:"driver"`
	User uint `json:"user"`
}


func (smsRequest *SmsRequest) Send() (error) {


	apiUrl := "https://www.bulksmsnigeria.com"
	resource := "/api/v1/sms/create/"
	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	req, _ := http.NewRequest("POST", urlStr, nil)
	data := req.URL.Query()
	data.Add("api_token", smsRequest.ApiToken)
	data.Add("to", smsRequest.To)
	data.Add("from", smsRequest.From)
	data.Add("body", smsRequest.Body)
	data.Add("dnd", "1")

	req.URL.RawQuery = data.Encode()
	cli := &http.Client{}
	response, err := cli.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(response.Status)
	return nil
}

func (request *MailRequest) Send() (error) {
	return SendEmail(request)
}

func SendEmail(request *MailRequest) error {

	from := mail.NewEmail("CitiKab", os.Getenv("email"))
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

