package deploycmd

import (
	"github.com/a-h/cwexport/cdk"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func Run(ms *types.MetricStat) error {
	app := awscdk.NewApp(nil)
	cdk.NewCDKStack(app, "cwexport", &cdk.CDKStackProps{}, ms)
	app.Synth(nil)
	return nil
}
