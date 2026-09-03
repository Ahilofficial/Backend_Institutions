package smtp

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(email string, subject string, body string) error {
	myemail := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	msg := fmt.Appendf(nil, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		myemail, email, subject, body)

	to := email
	address := host + ":" + port
	auth := smtp.PlainAuth("", myemail, password, host)
	err := smtp.SendMail(address, auth, myemail, []string{to}, msg)
	if err != nil {
		return err
	}
	return nil
}
