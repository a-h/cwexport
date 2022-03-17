package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/a-h/cwexport/db"
	"github.com/a-h/cwexport/deploycmd"
	"github.com/a-h/cwexport/firehose"
	"github.com/a-h/cwexport/processor"
	"github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	zc := zap.NewProductionConfig()
	zc.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logger, _ := zc.Build()
	defer logger.Sync()

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to load test aws config %w", err))
	}

	fh, err := firehose.New("cwexport-MetricDeliveryStreamA287D02E-oK0s7aoP5Bi1", cfg)
	if err != nil {
		logger.Fatal("Cannot create firehose", zap.Error(err))
		return
	}

	store, err := db.NewMetricStore("cwexport-CWExportMetricTable116C3288-14FQXYEUTR25R", cfg.Region)
	if err != nil {
		logger.Fatal("Cannot create store", zap.Error(err))
		return
	}

	m := cw.Metric{
		Namespace:   "authApi",
		Name:        "challengesStarted",
		ServiceName: "auth-api-challengePostHandler92AD93BF-UH40AniBZd25",
		ServiceType: "AWS::Lambda::Function",
	}

	start := time.Date(2022, time.March, 15, 9, 00, 0, 0, time.UTC)

	p, err := processor.New(logger, fh, store)
	if err != nil {
		logger.Error("Failed to create new processor", zap.Error(err))
		return
	}

	err = p.Process(ctx, start, m)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func deployCmd(args []string) {
	cmd := flag.NewFlagSet("deploy", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	configFlag := cmd.String("config", "", "Config file")

	var messages []string

	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}

	if *configFlag == "" {
		messages = append(messages, "Missing config file")
	}

	confData, err := ioutil.ReadFile(*configFlag)
	if err != nil {
		messages = append(messages, "Unable to read config")
	}

	var ms cw.MetricStat
	_, err = toml.Decode(string(confData), &ms)
	if err != nil {
		messages = append(messages, "Unable to parse config file")
	}

	if len(messages) > 0 {
		fmt.Println("Errors:")
		for _, m := range messages {
			fmt.Printf("  %s\n", m)
		}
		os.Exit(1)
	}

	err = deploycmd.Run(&ms)
	if err != nil {
		logger.Error("An error occured during processing", zap.Error(err))
		return
	}
}
