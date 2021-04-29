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

type logLevel int

const (
	logLevelTrace logLevel = iota
	logLevelDebug
	logLevelInfo
	logLevelWarning
	logLevelError
	logLevelFatal
)

func (ll logLevel) String() string {
	switch ll {
	case logLevelTrace:
		return "TRACE"
	case logLevelDebug:
		return "DEBUG"
	case logLevelInfo:
		return "INFO"
	case logLevelWarning:
		return "WARNING"
	case logLevelError:
		return "ERROR"
	case logLevelFatal:
		return "FATAL"
	default:
		panic(fmt.Sprintf("Invalid log level: %d", ll))
	}
}

type logger func(logLevel, interface{})

func (l logger) trace(v interface{}) {
	l(logLevelTrace, v)
}

func (l logger) debug(v interface{}) {
	l(logLevelDebug, v)
}

func (l logger) info(v interface{}) {
	l(logLevelInfo, v)
}

func (l logger) warning(v interface{}) {
	l(logLevelWarning, v)
}

func (l logger) error(v interface{}) {
	l(logLevelError, v)
}

func (l logger) rsp(
	status int,
	body string,
	logFormat string,
	v ...interface{},
) events.APIGatewayV2HTTPResponse {
	l.info(struct {
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

func (l logger) response(
	status int,
	body string,
) events.APIGatewayV2HTTPResponse {
	return l.rsp(status, body, "")
}

func (l logger) badRequest(body string) events.APIGatewayV2HTTPResponse {
	return l.response(http.StatusBadRequest, body)
}

func (l logger) internalServerError(
	format string,
	v ...interface{},
) events.APIGatewayV2HTTPResponse {
	return l.rsp(http.StatusInternalServerError, "", format, v...)
}

func requestLogger(requestID string) logger {
	return func(level logLevel, v interface{}) {
		logJSON(struct {
			RequestID string      `json:"request_id"`
			Level     logLevel    `json:"log_level"`
			Body      interface{} `json:"body"`
		}{
			RequestID: requestID,
			Level:     level,
			Body:      v,
		})
	}
}
