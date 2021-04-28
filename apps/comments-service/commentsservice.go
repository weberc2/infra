package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

type Comment struct {
	ID        string `json:"id"`
	PostID    string `json:"post_id"`
	Timestamp string `json:"timestamp"`
	Username  string `json:"username"`
	Body      string `json:"body"`
}

type CommentsService struct {
	DB    *dynamodb.DynamoDB
	Table string
}

func (cs *CommentsService) PutComment(comment *Comment) error {
	item, err := dynamodbattribute.MarshalMap(Comment{
		ID:        uuid.NewString(),
		PostID:    "some-post.html",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Username:  "weberc2",
		Body:      "Hello, world!",
	})
	if err != nil {
		return fmt.Errorf("Marshaling comment into dynamodb item: %w", err)
	}
	_, err = cs.DB.PutItem(&dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(cs.Table),
	})
	return err
}

func (cs *CommentsService) PostComments(postID string) ([]Comment, error) {
	expr, err := expression.NewBuilder().WithFilter(expression.Name("post_id").Equal(expression.Value(postID))).Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to build post-comments expression: %v", err))
	}
	out, err := cs.DB.Query(&dynamodb.QueryInput{
		KeyConditionExpression:    expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(cs.Table),
	})
	if err != nil {
		return nil, fmt.Errorf("Querying comments: %w", err)
	}

	comments := make([]Comment, len(out.Items))
	for i, item := range out.Items {
		if err := dynamodbattribute.ConvertFromMap(
			item,
			&comments[i],
		); err != nil {
			return nil, fmt.Errorf(
				"Failed to unmarshal dynamodb item into a comment: %w",
				err,
			)
		}
	}
	return comments, nil
}
