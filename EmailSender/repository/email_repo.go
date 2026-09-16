package repository

import (
	"backend_institutions/EmailSender/smtp"
)

type EmailRepository struct{}

func NewEmailRepository() *EmailRepository {
	return &EmailRepository{}
}

func (r *EmailRepository) SendMail(email, subject, body string) error {
	err := smtp.SendEmail(email, subject, body)

	if err != nil {
		return err
	}

	return nil
}
