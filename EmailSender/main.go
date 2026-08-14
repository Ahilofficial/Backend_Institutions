package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"backend_institutions/EmailSender/notificationpb"
	"backend_institutions/EmailSender/wire"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("Cant able to load the envirornment variable")
	}

	notificationService, err := wire.InitializeNotificationService()
	if err != nil {
		log.Fatalf("Failed to initialize notification service: %v", err)
	}

	grpcServer := grpc.NewServer()

	notificationpb.RegisterSendMailServer(
		grpcServer,
		notificationService,
	)

	port := os.Getenv("NOTIFICATION_GRPC_PORT")
	if port == "" {
		port = "15052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Notification gRPC Server started on %s", port)

	grpcServer.Serve(lis)
}
