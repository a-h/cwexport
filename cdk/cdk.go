package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	firehose "github.com/aws/aws-cdk-go/awscdkkinesisfirehosealpha/v2"
	destinations "github.com/aws/aws-cdk-go/awscdkkinesisfirehosedestinationsalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CDKStackProps struct {
	awscdk.StackProps
}

func NewCDKStack(scope constructs.Construct, id string, props *CDKStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create a bucket to store the output metrics.
	mob := awss3.NewBucket(stack, jsii.String("MetricOutput"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		EnforceSSL:        jsii.Bool(true),
		Versioned:         jsii.Bool(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
	})
	mob.AddLifecycleRule(&awss3.LifecycleRule{
		AbortIncompleteMultipartUploadAfter: awscdk.Duration_Days(jsii.Number(7)), // Save space by clearing up partial uploads.
		NoncurrentVersionExpiration:         awscdk.Duration_Days(jsii.Number(7)), // Delete old versions after 7 days.
		Expiration:                          awscdk.Duration_Days(jsii.Number(7)), // Delete files from pipeline bucket after 7 days.
	})
	fh := firehose.NewDeliveryStream(stack, jsii.String("MetricDeliveryStream"), &firehose.DeliveryStreamProps{
		Destinations: &[]firehose.IDestination{
			destinations.NewS3Bucket(mob, &destinations.S3BucketProps{
				BufferingInterval: awscdk.Duration_Minutes(jsii.Number(1.0)),
				BufferingSize:     awscdk.Size_Mebibytes(jsii.Number(5.0)),
				DataOutputPrefix:  jsii.String("cwexport"),
				ErrorOutputPrefix: jsii.String("cwexport_failures"),
			}),
		},
		Encryption: firehose.StreamEncryption_AWS_OWNED,
	})
	awscdk.NewCfnOutput(stack, jsii.String("Firehose"), &awscdk.CfnOutputProps{
		ExportName: jsii.String("Firehose"),
		Value:      fh.DeliveryStreamName(),
	})

	db := awsdynamodb.NewTable(stack, jsii.String("CWExportMetricTable"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("_pk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("_sk"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:         awsdynamodb.BillingMode_PAY_PER_REQUEST,
		Encryption:          awsdynamodb.TableEncryption_AWS_MANAGED,
		PointInTimeRecovery: jsii.Bool(true),
		RemovalPolicy:       awscdk.RemovalPolicy_DESTROY,
		TimeToLiveAttribute: jsii.String("_ttl"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("CWMetricTableOutput"), &awscdk.CfnOutputProps{
		ExportName: jsii.String("CWTableName"),
		Value:      db.TableName(),
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)
	NewCDKStack(app, "cwexport", &CDKStackProps{})
	app.Synth(nil)
}
