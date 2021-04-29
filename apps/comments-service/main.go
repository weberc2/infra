package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

func main() {
	commentsService := CommentsService{
		DB:    dynamodb.New(session.Must(session.NewSession())),
		Table: os.Getenv("DYNAMO_TABLE_NAME"),
	}

	lambda.Start(func(
		ctx context.Context,
		event events.APIGatewayV2HTTPRequest,
	) (events.APIGatewayV2HTTPResponse, error) {
		log := requestLogger(event.RequestContext.RequestID)
		log.Info(map[string]interface{}{
			"message": "Received new request",
			"http":    event.RequestContext.HTTP,
		})

		if event.RequestContext.Authorizer == nil {
			return log.InternalServerError("Missing authorizer information"), nil
		}

		if event.RequestContext.Authorizer.JWT == nil {
			return log.InternalServerError(
				"Wrong authorizer type: missing JWT authorizer information",
			), nil
		}

		username, ok := event.RequestContext.Authorizer.JWT.Claims["username"]
		if !ok {
			return log.InternalServerError("Missing username claim"), nil
		}

		var comment struct {
			PostID string `json:"post_id"`
			Body   string `json:"body"`
		}

		if err := json.Unmarshal([]byte(event.Body), &comment); err != nil {
			return log.BadRequest(fmt.Sprintf(
				"Error unmarshalling comment: %v",
				err,
			)), nil
		}

		if err := commentsService.PutComment(&Comment{
			ID:        uuid.NewString(),
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			PostID:    comment.PostID,
			Username:  username,
			Body:      comment.Body,
		}); err != nil {
			return log.InternalServerError(
				"Writing comment to database: %v",
				err,
			), nil
		}

		return log.Response(http.StatusCreated, ""), nil
	})
}
