package grpc

import (
	"backend_institutions/EmailSender/notificationpb"
	"context"
	"log"
	"os"

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
	log.Printf("Connected to Notification Service on port %s\n", port)
	return nil
}

func SendEmail(email string, subject string, body string, check string) error {
	_, err := Notificationclient.SendMail(
		context.Background(),
		&notificationpb.MailRequest{
			To:      email,
			Subject: subject,
			Body:    body,
			Check:   check,
		},
	)
	return err
}