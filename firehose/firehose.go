package firehose

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsfirehose "github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

type Firehose struct {
	DeliveryStreamName string
	FirehoseClient     *awsfirehose.Client
}

func New(config aws.Config, deliveryStreamName string) (fh Firehose, err error) {
	return Firehose{
		DeliveryStreamName: deliveryStreamName,
		FirehoseClient:     awsfirehose.NewFromConfig(config),
	}, nil
}

func (f Firehose) Put(ctx context.Context, metrics []interface{}) error {
	if len(metrics) == 0 {
		return nil
	}
	records := make([]types.Record, len(metrics))
	for i := 0; i < len(metrics); i++ {
		data, err := json.Marshal(metrics[i])
		if err != nil {
			return err
		}
		records[i] = types.Record{
			Data: data,
		}
	}

	_, err := f.FirehoseClient.PutRecordBatch(ctx, &awsfirehose.PutRecordBatchInput{
		DeliveryStreamName: &f.DeliveryStreamName,
		Records:            records,
	})
	return err
}
