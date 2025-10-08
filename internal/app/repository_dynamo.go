package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type dynamoContentRepository struct {
	client   dynamodbiface.DynamoDBAPI
	table    string
	scanPage int64
}

func init() {
	repo, err := newDynamoContentRepositoryFromEnv()
	if err != nil {
		log.Printf("helloworld: DynamoDB repository not initialised, falling back to in-memory store: %v", err)
		return
	}
	setContentRepository(repo)
}

func newDynamoContentRepositoryFromEnv() (ContentRepository, error) {
	table := os.Getenv("DYNAMODB_TABLE")
	if table == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE environment variable not set")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION or AWS_DEFAULT_REGION environment variable not set")
	}

	roleArn := os.Getenv("AWS_ROLE_ARN")
	if roleArn == "" {
		return nil, fmt.Errorf("AWS_ROLE_ARN environment variable not set")
	}

	sessionName := os.Getenv("AWS_ROLE_SESSION_NAME")
	if sessionName == "" {
		sessionName = fmt.Sprintf("go-helloworld-%d", time.Now().Unix())
	}

	baseSession, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	var creds *credentials.Credentials
	if tokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"); tokenFile != "" {
		creds = stscreds.NewWebIdentityCredentials(baseSession, roleArn, sessionName, tokenFile)
	} else {
		creds = stscreds.NewCredentials(baseSession, roleArn, func(p *stscreds.AssumeRoleProvider) {
			p.RoleSessionName = sessionName
		})
	}

	if _, err := creds.Get(); err != nil {
		return nil, fmt.Errorf("failed to obtain STS credentials: %w", err)
	}

	client := dynamodb.New(baseSession, aws.NewConfig().WithCredentials(creds))

	return &dynamoContentRepository{
		client:   client,
		table:    table,
		scanPage: 25,
	}, nil
}

func (r *dynamoContentRepository) ListContent(ctx context.Context) (allContent, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.table),
		Limit:     aws.Int64(r.scanPage),
	}

	result := make(allContent, 0)
	var innerErr error

	err := r.client.ScanPagesWithContext(ctx, input, func(page *dynamodb.ScanOutput, lastPage bool) bool {
		for _, item := range page.Items {
			content, convErr := dynamoItemToContent(item)
			if convErr != nil {
				innerErr = convErr
				return false
			}
			result = append(result, *content)
		}
		return true
	})
	if innerErr != nil {
		return nil, innerErr
	}
	if err != nil {
		return nil, fmt.Errorf("scan DynamoDB: %w", err)
	}
	return result, nil
}

func (r *dynamoContentRepository) GetContent(ctx context.Context, id string) (*api, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
		ConsistentRead: aws.Bool(true),
	}

	out, err := r.client.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("get item from DynamoDB: %w", err)
	}
	if out.Item == nil {
		return nil, ErrContentNotFound
	}
	content, convErr := dynamoItemToContent(out.Item)
	if convErr != nil {
		return nil, convErr
	}
	return content, nil
}

func (r *dynamoContentRepository) CreateContent(ctx context.Context, item api) (*api, error) {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.table),
		Item: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(item.ID),
			},
			"name": {
				S: aws.String(item.Name),
			},
		},
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	}

	if _, err := r.client.PutItemWithContext(ctx, input); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return nil, ErrContentAlreadyExists
		}
		return nil, fmt.Errorf("put item into DynamoDB: %w", err)
	}

	copy := item
	return &copy, nil
}

func (r *dynamoContentRepository) UpdateContent(ctx context.Context, id string, name string) (*api, error) {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
		UpdateExpression: aws.String("SET #n = :name"),
		ExpressionAttributeNames: map[string]*string{
			"#n": aws.String("name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name": {S: aws.String(name)},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
		ReturnValues:        aws.String(dynamodb.ReturnValueAllNew),
	}

	out, err := r.client.UpdateItemWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return nil, ErrContentNotFound
		}
		return nil, fmt.Errorf("update item in DynamoDB: %w", err)
	}

	content, convErr := dynamoItemToContent(out.Attributes)
	if convErr != nil {
		return nil, convErr
	}
	return content, nil
}

func (r *dynamoContentRepository) DeleteContent(ctx context.Context, id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.table),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	}

	if _, err := r.client.DeleteItemWithContext(ctx, input); err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return ErrContentNotFound
		}
		return fmt.Errorf("delete item from DynamoDB: %w", err)
	}
	return nil
}

func dynamoItemToContent(item map[string]*dynamodb.AttributeValue) (*api, error) {
	idAttr, ok := item["id"]
	if !ok || idAttr.S == nil {
		return nil, fmt.Errorf("dynamodb item missing id attribute")
	}
	nameAttr, ok := item["name"]
	if !ok || nameAttr.S == nil {
		return nil, fmt.Errorf("dynamodb item missing name attribute")
	}
	return &api{ID: aws.StringValue(idAttr.S), Name: aws.StringValue(nameAttr.S)}, nil
}
