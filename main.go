package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/a-h/cwexport/cw"
)

func main() {
	m := cw.Metric{
		Namespace:   "authApi",
		Name:        "challengesStarted",
		ServiceName: "auth-api-challengePostHandler92AD93BF-UH40AniBZd25",
		ServiceType: "AWS::Lambda::Function",
		StartTime:   time.Date(2022, time.March, 14, 16, 00, 0, 0, time.UTC),
		EndTime:     time.Date(2022, time.March, 14, 17, 30, 0, 0, time.UTC),
	}

	metrics, err := cw.GetMetrics(m)
	if err != nil {
		log.Fatal(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", " ")
	err = e.Encode(metrics)
	if err != nil {
		log.Fatal(err)
	}
}
