package middleware

import (
	"backend_institutions/internal/grpc"
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func RequestResponseLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		method := c.Method()
		endpoint := c.Path()
		reqBody := string(c.Body())

		err := c.Next()

		status := c.Response().StatusCode()
		respBody := string(c.Response().Body())

		cleanEndpoint := strings.ToLower(endpoint)
		if strings.Contains(cleanEndpoint, "signup") || strings.Contains(cleanEndpoint, "signin") || strings.Contains(cleanEndpoint, "password") {
			reqBody = "[REDACTED/SENSITIVE]"
		}

		go func() {
			sendLogErr := grpc.SendLog(
				"MainAPI",
				method,
				endpoint,
				reqBody,
				respBody,
				int32(status),
			)
			if sendLogErr != nil {
				log.Printf("Error sending request/response log to gRPC service: %v\n", sendLogErr)
			}
		}()

		return err
	}
}
