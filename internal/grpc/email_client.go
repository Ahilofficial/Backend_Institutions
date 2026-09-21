package grpc

import (
	"backend_institutions/EmailSender/notificationpb"
	"context"
	"errors"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Notificationclient notificationpb.SendMailClient
	NotificationConn   *grpc.ClientConn
)

func ConnectService() error {
	host := os.Getenv("NOTIFICATION_GRPC_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("NOTIFICATION_GRPC_PORT")
	if port == "" {
		port = "15052"
	}

	conn, err := grpc.NewClient(
		host+":"+port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	NotificationConn = conn
	Notificationclient = notificationpb.NewSendMailClient(conn)

	log.Printf(
		"Connected to Notification Service at %s:%s",
		host,
		port,
	)

	return nil
}

func SendEmail(
	email string,
	subject string,
	body string,
	check string,
) error {
	if Notificationclient == nil {
		if err := ConnectService(); err != nil || Notificationclient == nil {
			return errors.New("notification gRPC client is not initialized")
		}
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	_, err := Notificationclient.SendMail(
		ctx,
		&notificationpb.MailRequest{
			To:      email,
			Subject: subject,
			Body:    body,
			Check:   check,
		},
	)

	if err != nil {
		log.Printf("failed to send email: %v", err)
		return err
	}

	log.Printf("email request sent to notification service: %s", email)

	return nil
}