package middleware

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func SetupLogger() fiber.Handler {
	accessLog, err := os.OpenFile(
		"./access.log",
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatalf("error opening access.log file: %v", err)
	}

	logger := logger.New(logger.Config{
		Format: `{
  "time": "${time}",
  "url": "${url}"
  "status": ${status},
  "method": "${method}",
  "path": "${path}",
  "request_body": "${body}",
  "response_body": "${resBody}"
}
`,
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "Asia/Kolkata",
		Stream:     accessLog,
	})

	return logger
}
