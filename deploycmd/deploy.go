package deploycmd

import (
	"os"
	"os/exec"

	"github.com/a-h/cwexport/cdk"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type Arguments struct {
	Stats            *[]types.MetricStat
	FirehoseRoleName string
	BucketName       string
}

func Run(args Arguments) error {
	app := awscdk.NewApp(nil)
	cdk.NewCDKStack(app, "cwexport", &cdk.CDKStackProps{
		Stats:            args.Stats,
		FirehoseRoleName: args.FirehoseRoleName,
		BucketName:       args.BucketName,
	})
	cxa := app.Synth(nil)
	com := exec.Command("cdk", "deploy", "--app="+*cxa.Directory(), "--require-approval=never")
	com.Stdin = os.Stdin
	com.Stdout = os.Stdout
	com.Stderr = os.Stderr
	err := com.Run()
	if err != nil {
		panic(err)
	}
	return nil
}
