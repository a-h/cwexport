package main

import (
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
	awslambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
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

	bundlingOptions := &awslambdago.BundlingOptions{
		GoBuildFlags: &[]*string{jsii.String(`-ldflags "-s -w"`)},
	}

	f := awslambdago.NewGoFunction(stack, jsii.String("Processor"), &awslambdago.GoFunctionProps{
		Environment: &map[string]*string{
			"METRIC_TABLE_NAME":    db.TableName(),
			"METRIC_FIREHOSE_NAME": fh.DeliveryStreamName(),
		},
		LogRetention: awslogs.RetentionDays_FIVE_MONTHS,
		MemorySize:   jsii.Number(1024),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Tracing:      awslambda.Tracing_ACTIVE,
		Entry:        jsii.String("../lambda/processor"),
		Bundling:     bundlingOptions,
		Runtime:      awslambda.Runtime_GO_1_X(),
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

	m := &cw.MetricStat{
		Metric: &cw.Metric{
			Namespace:  aws.String("authApi"),
			MetricName: aws.String("challengesStarted"),
			Dimensions: []cw.Dimension{
				{
					Name:  aws.String("ServiceName"),
					Value: aws.String("auth-api-challengePostHandler92AD93BF-UH40AniBZd25"),
				},
				{
					Name:  aws.String("ServiceType"),
					Value: aws.String("AWS::Lambda::Function"),
				},
			},
		},
		Period: aws.Int32(5),
		Stat:   aws.String("Sum"),
	}

	awsevents.NewRule(stack, jsii.String("Scheduler"), &awsevents.RuleProps{
		Schedule: awsevents.Schedule_Rate(awscdk.Duration_Minutes(jsii.Number(1))),
		Targets: &[]awsevents.IRuleTarget{
			awseventstargets.NewLambdaFunction(f, &awseventstargets.LambdaFunctionProps{
				Event: awsevents.RuleTargetInput_FromObject(m),
			}),
		},
	})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)
	NewCDKStack(app, "cwexport", &CDKStackProps{})
	app.Synth(nil)
}
