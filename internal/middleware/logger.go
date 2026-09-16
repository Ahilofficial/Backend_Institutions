package middleware

import (
	"time"

	"backend_institutions/internal/grpc"
	"backend_institutions/internal/grpc/loggerpb"

	"github.com/gofiber/fiber/v3"
)

func RequestResponseLogger() fiber.Handler {

	return func(c fiber.Ctx) error {

		start := time.Now()

		method := c.Method()
		endpoint := c.OriginalURL()
		ip := c.IP()

		requestBody := string(c.Body())

		err := c.Next()

		status := c.Response().StatusCode()
		responseBody := string(c.Response().Body())

		duration := time.Since(start)

		errorMessage := ""

		if err != nil {
			errorMessage = err.Error()
		}

		grpc.SendLog(&loggerpb.LogRequest{
			Time:         time.Now().Format("2006-01-02 15:04:05"),
			Service:      "institution-service",
			Method:       method,
			Endpoint:     endpoint,
			Status:       int32(status),
			Latency:      duration.String(),
			Ip:           ip,
			RequestBody:  requestBody,
			ResponseBody: responseBody,
			Error:        errorMessage,
		})

		return err
	}
}