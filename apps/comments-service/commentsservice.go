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

// Comment represents a comment.
type Comment struct {
	// ID identifies a comment within a particular site.
	ID string `json:"id"`

	// PostID identifies the post with which the comment is associated.
	PostID string `json:"post_id"`

	// Timestamp marks the time at which the comment was originally published.
	// The format is ISO-8601.  The equivalent Go `time` layout string is
	// `time.RFC3339Nano`.
	Timestamp string `json:"timestamp"`

	// Username identifies the user associated with the comment.
	Username string `json:"username"`

	// Body holds the body of the comment.
	Body string `json:"body"`
}

// CommentsService is a client for CRUDing comments.
type CommentsService struct {
	// DB is a dynamodb client.
	DB *dynamodb.DynamoDB

	// Table is the name of the dynamodb comments table.
	Table string
}

// PutComment inserts a comment into the comments table.
func (cs *CommentsService) PutComment(comment *Comment) error {
	item, err := dynamodbattribute.MarshalMap(Comment{
		ID:        uuid.NewString(),
		PostID:    "some-post.html",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Username:  "weberc2",
		Body:      "Hello, world!",
	})
	if err != nil {
		return fmt.Errorf("marshaling comment into dynamodb item: %w", err)
	}
	_, err = cs.DB.PutItem(&dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(cs.Table),
	})
	return err
}

// PostComments retrieves the comments associated with a particular post.
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
		return nil, fmt.Errorf("querying comments: %w", err)
	}

	comments := make([]Comment, len(out.Items))
	for i, item := range out.Items {
		if err := dynamodbattribute.UnmarshalMap(
			item,
			&comments[i],
		); err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal dynamodb item into a comment: %w",
				err,
			)
		}
	}
	return comments, nil
}
