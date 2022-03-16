package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	h := NewHandler(log)
	lambda.Start(h.Handle)
}

func NewHandler(logger *zap.Logger) Handler {
	return Handler{
		logger: logger,
	}
}

type Handler struct {
	logger *zap.Logger
}

func (h Handler) Handle(ctx context.Context, event events.CloudWatchEvent) (err error) {
	h.logger.Info("Received event", zap.Any("event", event))
	//TODO: Create a processor, possibly using dependenices of the handler.
	//TODO: Call the process method.
	return nil
}
