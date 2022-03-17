package deploycmd

import (
	"github.com/a-h/cwexport/cdk"
	"github.com/aws/aws-cdk-go/awscdk/v2"
)

func Run() error {
	app := awscdk.NewApp(nil)
	cdk.NewCDKStack(app, "cwexport", &cdk.CDKStackProps{})
	app.Synth(nil)
	return nil
}
