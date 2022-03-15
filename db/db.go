package db

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDBDescribeTableAPI defines the interface for the DescribeTable function.
// We use this interface to enable unit testing.
type DynamoDBDescribeTableAPI interface {
	DescribeTable(ctx context.Context,
		params *dynamodb.DescribeTableInput,
		optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
}

func GetTableInfo(c context.Context, api DynamoDBDescribeTableAPI, input *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
	return api.DescribeTable(c, input)
}

func DescribeTable() {
	client, _ := Connect()
	table := "Test"

	// Build the input parameters for the request.
	input := &dynamodb.DescribeTableInput{
		TableName: &table,
	}

	resp, err := GetTableInfo(context.TODO(), client, input)
	if err != nil {
		panic("failed to describe table, " + err.Error())
	}

	fmt.Println("Info about " + table + ":")
	fmt.Println("  #items:     ", resp.Table.ItemCount)
	fmt.Println("  Size (bytes)", resp.Table.TableSizeBytes)
	fmt.Println("  Status:     ", string(resp.Table.TableStatus))
}

func Connect() (*dynamodb.Client, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithEndpointResolverWithOptions(
		aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: "http://localhost:8000"}, nil
		}),
	),
	)

	if err != nil {
		err = fmt.Errorf("unable to load SDK config: %w", err)
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}
