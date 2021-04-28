package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func logJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error while logging `%#v` as JSON: %v", v, err)
	}
	log.Printf("%s", data)
}

type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelFatal
)

func (ll LogLevel) String() string {
	switch ll {
	case LogLevelTrace:
		return "TRACE"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		panic(fmt.Sprintf("Invalid log level: %d", ll))
	}
}

type Logger func(LogLevel, interface{})

func (l Logger) Trace(v interface{}) {
	l(LogLevelTrace, v)
}

func (l Logger) Debug(v interface{}) {
	l(LogLevelDebug, v)
}

func (l Logger) Info(v interface{}) {
	l(LogLevelInfo, v)
}

func (l Logger) Warning(v interface{}) {
	l(LogLevelWarning, v)
}

func (l Logger) Error(v interface{}) {
	l(LogLevelError, v)
}

func (l Logger) response(
	status int,
	body string,
	logFormat string,
	v ...interface{},
) events.APIGatewayV2HTTPResponse {
	l.Info(struct {
		StatusCode int    `json:"status_code"`
		StatusText string `json:"status_text"`
		Body       string `json:"body"`
		Message    string `json:"message,omitempty"`
	}{
		StatusCode: status,
		StatusText: http.StatusText(status),
		Body:       body,
		Message:    fmt.Sprintf(logFormat, v...),
	})

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      status,
		Body:            body,
		IsBase64Encoded: false,
	}
}

func (l Logger) Response(
	status int,
	body string,
) events.APIGatewayV2HTTPResponse {
	return l.response(status, body, "")
}

func (l Logger) BadRequest(body string) events.APIGatewayV2HTTPResponse {
	return l.Response(http.StatusBadRequest, body)
}

func (l Logger) InternalServerError(
	format string,
	v ...interface{},
) events.APIGatewayV2HTTPResponse {
	return l.response(http.StatusInternalServerError, "", format, v...)
}

func requestLogger(requestID string) Logger {
	return func(level LogLevel, v interface{}) {
		logJSON(struct {
			RequestID string      `json:"request_id"`
			Level     LogLevel    `json:"log_level"`
			Body      interface{} `json:"body"`
		}{
			RequestID: requestID,
			Level:     level,
			Body:      v,
		})
	}
}
