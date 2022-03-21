package cdk

import (
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	firehose "github.com/aws/aws-cdk-go/awscdkkinesisfirehosealpha/v2"
	destinations "github.com/aws/aws-cdk-go/awscdkkinesisfirehosedestinationsalpha/v2"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

//go:embed lambda/processor/lambda
var lambdaBinary embed.FS

type CDKStackProps struct {
	// Stats to record.
	Stats *[]types.MetricStat
	// FirehoseRoleName allows a custom role to be used for the Firehose. If left empty, a new role will be created.
	FirehoseRoleName string
	// BucketName is an optional bucket name to use as a target. If left empty, a new bucket will be created.
	BucketName string
	awscdk.StackProps
}

func NewCDKStack(scope constructs.Construct, id string, props *CDKStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	mob := getOrCreateBucket(stack, props.BucketName)

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

	dir, err := ioutil.TempDir("", "cwexport")
	if err != nil {
		panic("Cannot create temporary directory: " + err.Error())
	}
	defer os.RemoveAll(dir)
	lf, err := lambdaBinary.Open("lambda/processor/lambda")
	if err != nil {
		panic("Cannot open lambda binary: " + err.Error())
	}
	tmp, err := os.Create(path.Join(dir, "lambda"))
	if err != nil {
		panic("Cannot create temporary file in directory: " + err.Error())
	}
	_, err = io.Copy(tmp, lf)
	if err != nil {
		panic("Cannot copy lambda binary to temporary location: " + err.Error())
	}

	var fhRole awsiam.IRole
	if props.FirehoseRoleName != "" {
		fhRole = awsiam.Role_FromRoleName(stack, jsii.String("CustomFirehoseRole"), &props.FirehoseRoleName)
	}

	for _, m := range *props.Stats {
		fh := firehose.NewDeliveryStream(stack, jsii.String(fmt.Sprintf("%s-%s-MetricDeliveryStream", *m.Metric.Namespace, *m.Metric.MetricName)), &firehose.DeliveryStreamProps{
			Destinations: &[]firehose.IDestination{
				destinations.NewS3Bucket(mob, &destinations.S3BucketProps{
					BufferingInterval: awscdk.Duration_Minutes(jsii.Number(1.0)),
					BufferingSize:     awscdk.Size_Mebibytes(jsii.Number(5.0)),
					DataOutputPrefix:  jsii.String(fmt.Sprintf("cwexport-%s-%s", *m.Metric.Namespace, *m.Metric.MetricName)),
					ErrorOutputPrefix: jsii.String(fmt.Sprintf("cwexport_failures-%s-%s", *m.Metric.Namespace, *m.Metric.MetricName)),
					Role:              fhRole,
				}),
			},
			Encryption: firehose.StreamEncryption_AWS_OWNED,
		})
		awscdk.NewCfnOutput(stack, jsii.String(fmt.Sprintf("%s-%s-Firehose", *m.Metric.Namespace, *m.Metric.MetricName)), &awscdk.CfnOutputProps{
			ExportName: jsii.String(fmt.Sprintf("%s-%s-Firehose", *m.Metric.Namespace, *m.Metric.MetricName)),
			Value:      fh.DeliveryStreamName(),
		})

		f := awslambda.NewFunction(stack, jsii.String(fmt.Sprintf("%s-%s-Processor", *m.Metric.Namespace, *m.Metric.MetricName)), &awslambda.FunctionProps{
			Environment: &map[string]*string{
				"METRIC_TABLE_NAME":    db.TableName(),
				"METRIC_FIREHOSE_NAME": fh.DeliveryStreamName(),
			},
			LogRetention: awslogs.RetentionDays_FIVE_MONTHS,
			Code:         awslambda.AssetCode_FromAsset(jsii.String(dir), nil),
			MemorySize:   jsii.Number(1024),
			Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
			Tracing:      awslambda.Tracing_ACTIVE,
			Runtime:      awslambda.Runtime_GO_1_X(),
			Handler:      jsii.String("lambda"),
			InitialPolicy: &[]awsiam.PolicyStatement{
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Actions:   jsii.Strings("cloudwatch:GetMetricData"),
					Effect:    awsiam.Effect_ALLOW,
					Resources: jsii.Strings("*"),
				}),
			},
		})
		db.GrantReadWriteData(f)
		fh.GrantPutRecords(f)

		awsevents.NewRule(stack, jsii.String(fmt.Sprintf("%s-%s-Scheduler", *m.Metric.Namespace, *m.Metric.MetricName)), &awsevents.RuleProps{
			Schedule: awsevents.Schedule_Rate(awscdk.Duration_Minutes(jsii.Number(5))),
			Targets: &[]awsevents.IRuleTarget{
				awseventstargets.NewLambdaFunction(f, &awseventstargets.LambdaFunctionProps{
					Event: awsevents.RuleTargetInput_FromObject(m),
				}),
			},
		})
	}

	return stack
}

func getOrCreateBucket(stack constructs.Construct, bucketName string) awss3.IBucket {
	if bucketName != "" {
		return awss3.Bucket_FromBucketName(stack, jsii.String("MetricOutput"), &bucketName)
	}
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
	return mob
}
