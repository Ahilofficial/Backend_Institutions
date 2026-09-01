package utilities

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func WriteAppLog(service, method, endpoint string, status int, request, response interface{}) error {
	path := "logs/app.log"

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		requestBody = []byte(fmt.Sprintf("%v", request))
	}
	responseBody, err := json.Marshal(response)
	if err != nil {
		responseBody = []byte(fmt.Sprintf("%v", response))
	}

	logContent := fmt.Sprintf(`
**
Time      : %s
Service   : %s
Method    : %s
Endpoint  : %s
Status    : %d

Request:
%s

Response:
%s

**

`,
		time.Now().Format("2006-01-02 15:04:05"),
		service,
		method,
		endpoint,
		status,
		string(requestBody),
		string(responseBody),
	)

	_, err = file.WriteString(logContent)
	if err != nil {
		fmt.Println("Cant able to write the data inside the log")
	} else {
		fmt.Println("Written the data successfully")
	}
	return err
}

func WriteEmailLog(to, subject string, success bool, errorMsg string) error {
	path := "logs/notification.log"

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err

	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	status := "SUCCESS"
	detail := ""
	if !success {
		status = "FAILED"
		detail = fmt.Sprintf(" | Error: %s", errorMsg)
	}

	logEntry := fmt.Sprintf("[%s] Status: %s | To: %s | Subject: %s%s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		status,
		to,
		subject,
		detail,
	)

	_, err = file.WriteString(logEntry)
	if err != nil {
		fmt.Println("Cant able to write the data inside the log")
	} else {
		fmt.Println("Written the data successfully")
	}
	return err
}
