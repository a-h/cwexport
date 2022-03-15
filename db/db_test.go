package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var testClient *dynamodb.Client

func init() {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}
	creds := credentials.NewStaticCredentialsProvider("fake", "accessKeyId", "secretKeyId")
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		panic(fmt.Errorf("failed to load test aws config %w", err))
	}
	testClient = dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL(endpoint)))
}

const region = "eu-pluto-1"

func createLocalTable(t *testing.T) (name string) {
	name = uuid.New().String()
	_, err := testClient.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("_pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("_sk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("gsi1"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("gsi2"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("_pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("_sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
		TableName:   aws.String(name),
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("gsi1"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("gsi1"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("_pk"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType:   types.ProjectionTypeInclude,
					NonKeyAttributes: []string{"phoneVerificationExpiry", "_sk"},
				},
			},
			{
				IndexName: aws.String("gsi2"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("gsi2"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("_pk"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType:   types.ProjectionTypeInclude,
					NonKeyAttributes: []string{"emailVerificationExpiry"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create local table: %v", err)
	}
	return
}

func deleteLocalTable(t *testing.T, name string) {
	_, err := testClient.DeleteTable(context.Background(), &dynamodb.DeleteTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		t.Fatalf("failed to delete table: %v", err)
	}
}
