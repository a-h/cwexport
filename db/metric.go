package db

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
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

func getPartitionKey(m *cw.MetricStat) string {
	var sb strings.Builder
	sb.WriteString(*m.Metric.Namespace)
	sb.WriteRune('/')
	for _, d := range m.Metric.Dimensions {
		sb.WriteString(*d.Name)
		sb.WriteRune('/')
		sb.WriteString(*d.Value)
		sb.WriteRune('/')
	}
	sb.WriteString(*m.Metric.MetricName)
	sb.WriteRune('/')
	sb.WriteString(*m.Stat)
	sb.WriteRune('/')
	sb.WriteString(strconv.FormatInt(int64(*m.Period), 10))
	return sb.String()
}

func getSortKeyPosition() string {
	return "position"
}

// _pk             _sk               lastStart                    name     email
// ns/logins/sum   position          2022-04-01T13:13:35.000Z
// ns/logins/sum   user                                           adrian   a@example.com

func (ms MetricStore) Get(ctx context.Context, m *cw.MetricStat) (lastStart time.Time, ok bool, err error) {
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

func (ms MetricStore) Put(ctx context.Context, m *cw.MetricStat, lastStart time.Time) error {
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
