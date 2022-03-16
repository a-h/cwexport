package db

import (
	"context"
	"fmt"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type OptionsFunc func(*MetricStore)

func WithClient(client *dynamodb.Client) func(*MetricStore) {
	return func(ms *MetricStore) {
		ms.db = client
	}
}

func NewMetricStore(tableName, region string, options ...OptionsFunc) (s *MetricStore, err error) {
	s = &MetricStore{
		tableName: tableName,
	}
	for _, o := range options {
		o(s)
	}
	if s.db == nil {
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
		if err != nil {
			return s, err
		}
		s.db = dynamodb.NewFromConfig(cfg)
	}
	return
}

type MetricStore struct {
	db        *dynamodb.Client
	tableName string
}

func getPartitionKey(m cw.Metric) string {
	return fmt.Sprintf("%s/%s/%s/%s", m.Namespace, m.Name, m.ServiceName, m.ServiceType)
}

func getSortKeyPosition() string {
	return "position"
}

// _pk             _sk               lastStart                    name     email
// ns/logins/sum   position          2022-04-01T13:13:35.000Z
// ns/logins/sum   user                                           adrian   a@example.com

func (ms MetricStore) Get(ctx context.Context, m cw.Metric) (lastStart time.Time, ok bool, err error) {
	gio, err := ms.db.GetItem(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"_pk": &types.AttributeValueMemberS{
				Value: getPartitionKey(m),
			},
			"_sk": &types.AttributeValueMemberS{
				Value: getSortKeyPosition(),
			},
		},
		TableName:      &ms.tableName,
		ConsistentRead: aws.Bool(true),
	})
	if err != nil || gio.Item == nil {
		return
	}
	// Get the item.
	lsv, ok := gio.Item["lastStart"]
	if !ok {
		return
	}
	// Check the type of the item.
	lsvs, ok := lsv.(*types.AttributeValueMemberS)
	if !ok {
		return
	}
	// Parse the attribute value as a date.
	lastStart, err = time.Parse(time.RFC3339, lsvs.Value)
	return
}

func (ms MetricStore) Put(ctx context.Context, m cw.Metric, lastStart time.Time) error {
	_, err := ms.db.PutItem(ctx, &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"_pk": &types.AttributeValueMemberS{
				Value: getPartitionKey(m),
			},
			"_sk": &types.AttributeValueMemberS{
				Value: getSortKeyPosition(),
			},
			"lastStart": &types.AttributeValueMemberS{
				Value: lastStart.Format(time.RFC3339),
			},
		},
		TableName: &ms.tableName,
	})
	if err != nil {
		return err
	}
	return err
}
