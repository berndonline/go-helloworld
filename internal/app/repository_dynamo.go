package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type dynamoContentRepository struct {
	client   *dynamodb.Client
	table    string
	scanPage int32
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

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	stsClient := sts.NewFromConfig(cfg)

	var provider aws.CredentialsProvider
	if tokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"); tokenFile != "" {
		provider = stscreds.NewWebIdentityRoleProvider(stsClient, roleArn, stscreds.IdentityTokenFile(tokenFile), func(o *stscreds.WebIdentityRoleOptions) {
			o.RoleSessionName = sessionName
		})
	} else {
		provider = stscreds.NewAssumeRoleProvider(stsClient, roleArn, func(o *stscreds.AssumeRoleOptions) {
			o.RoleSessionName = sessionName
		})
	}

	creds := aws.NewCredentialsCache(provider)
	if _, err := creds.Retrieve(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to obtain STS credentials: %w", err)
	}
	cfg.Credentials = creds

	client := dynamodb.NewFromConfig(cfg)

	return &dynamoContentRepository{
		client:   client,
		table:    table,
		scanPage: 25,
	}, nil
}

func (r *dynamoContentRepository) ListContent(ctx context.Context) (allContent, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.table),
		Limit:     aws.Int32(r.scanPage),
	}

	result := make(allContent, 0)
	paginator := dynamodb.NewScanPaginator(r.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("scan DynamoDB: %w", err)
		}
		for _, item := range page.Items {
			content, convErr := dynamoItemToContent(item)
			if convErr != nil {
				return nil, convErr
			}
			result = append(result, *content)
		}
	}
	return result, nil
}

func (r *dynamoContentRepository) GetContent(ctx context.Context, id string) (*api, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.table),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ConsistentRead: aws.Bool(true),
	}

	out, err := r.client.GetItem(ctx, input)
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
		Item: map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: item.ID},
			"name": &types.AttributeValueMemberS{Value: item.Name},
		},
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	}

	if _, err := r.client.PutItem(ctx, input); err != nil {
		var conditionalErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalErr) {
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
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET #n = :name"),
		ExpressionAttributeNames: map[string]string{
			"#n": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: name},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
		ReturnValues:        types.ReturnValueAllNew,
	}

	out, err := r.client.UpdateItem(ctx, input)
	if err != nil {
		var conditionalErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalErr) {
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
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	}

	if _, err := r.client.DeleteItem(ctx, input); err != nil {
		var conditionalErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalErr) {
			return ErrContentNotFound
		}
		return fmt.Errorf("delete item from DynamoDB: %w", err)
	}
	return nil
}

func dynamoItemToContent(item map[string]types.AttributeValue) (*api, error) {
	idAttr, ok := item["id"]
	if !ok {
		return nil, fmt.Errorf("dynamodb item missing id attribute")
	}
	idValue, ok := idAttr.(*types.AttributeValueMemberS)
	if !ok {
		return nil, fmt.Errorf("dynamodb item id attribute is not a string")
	}
	nameAttr, ok := item["name"]
	if !ok {
		return nil, fmt.Errorf("dynamodb item missing name attribute")
	}
	nameValue, ok := nameAttr.(*types.AttributeValueMemberS)
	if !ok {
		return nil, fmt.Errorf("dynamodb item name attribute is not a string")
	}
	return &api{ID: idValue.Value, Name: nameValue.Value}, nil
}
