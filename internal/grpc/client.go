package grpc

import (
	"context"
	"log"
	"time"

	"backend_institutions/internal/grpc/loggerpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	LoggerConn   *grpc.ClientConn
	LoggerClient loggerpb.LoggerServiceClient
)

func ConnectLogger() error {

	conn, err := grpc.NewClient(
		"localhost:15051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	LoggerConn = conn
	LoggerClient = loggerpb.NewLoggerServiceClient(conn)

	log.Println("Logger gRPC client connected on :15051")

	return nil
}

func CloseLogger() {

	if LoggerConn != nil {
		LoggerConn.Close()
		log.Println("Logger gRPC connection closed")
	}
}

func SendLog(req *loggerpb.LogRequest) {

	if LoggerClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	_, err := LoggerClient.LogRequestResponse(ctx, req)

	if err != nil {
		log.Println("Logger gRPC error:", err)
	}
}