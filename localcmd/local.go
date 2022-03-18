package localcmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/a-h/cwexport/processor"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"go.uber.org/zap"
)

type Format string

const (
	FormatCSV  Format = "csv"
	FormatJSON        = "json"
)

type Args struct {
	Start      time.Time
	MetricStat *types.MetricStat
}

type nopMetricStore struct {
}

func (_ nopMetricStore) Get(ctx context.Context, m *types.MetricStat) (lastStart time.Time, ok bool, err error) {
	return
}
func (_ nopMetricStore) Put(ctx context.Context, m *types.MetricStat, lastStart time.Time) (err error) {
	return
}

type csvPutter struct {
	writer csv.Writer
}

func (p csvPutter) Put(ctx context.Context, ms []processor.MetricSample) error {
	for _, s := range ms {
		//TODO: Add a header if it's the first time?
		var record []string
		record = append(record, *s.Metric.Namespace)
		for _, dim := range s.Metric.Dimensions {
			record = append(record, fmt.Sprintf("%s/%s", *dim.Name, *dim.Value))
		}
		record = append(record, *s.Metric.MetricName)
		record = append(record, *s.Stat)
		record = append(record, s.Sample.Time.Format(time.RFC3339))
		record = append(record, fmt.Sprintf("%f", s.Sample.Value))
		err := p.writer.Write(record)
		if err != nil {
			return err
		}
	}
	return nil
}

func Run(args Args) (err error) {
	logger := zap.NewNop()
	csvp := csvPutter{
		writer: *csv.NewWriter(os.Stdout),
	}
	defer csvp.writer.Flush()
	p, err := processor.New(logger, nopMetricStore{}, csvp.Put, cw.Cloudwatch{})
	if err != nil {
		logger.Error("Failed to create new processor", zap.Error(err))
		return
	}

	err = p.Process(context.Background(), args.Start, args.MetricStat)
	if err != nil {
		logger.Error("An error occured during processing", zap.Error(err))
		return
	}
	return nil
}
