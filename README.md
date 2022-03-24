# cwexport

Exports CloudWatch metrics to JSON or CSV.

Can be used as a local command-line tool, or deployed as a Lambda function.

## Usage

### Local export (CSV)

```sh
./cwexport local \
  -from=2022-03-14T16:00:00Z \
  -ns=authApi \
  -name=challengesStarted \
  -stat=Sum \
  -dimension=ServiceName/auth-api-challengePostHandler92AD93BF-thIg6mklFAlF \
  -dimension=ServiceType/AWS::Lambda::Function
```

### Lambda export to S3 (JSON)

```sh
./cwexport deploy \
  -config=test-config.toml
```

```yaml
[[metric]]
Namespace="pricingApi"
MetricName="completedPricing"
Stat="Sum"
Period = 5
StartTime=2021-03-21T09:00:00Z
[metric.dimensions]
ServiceName="pricing-api-generatePricePost3D9C7A9C-UERajJf7MzlI"
ServiceType="AWS::Lambda::Function"

[[metric]]
Namespace="AWS/Lambda"
MetricName="Invocations"
Stat="Sum"
Period=5
StartTime=2021-03-21T09:00:00Z
```

## Tasks

### run

Requires: build

Run locally & outputs CSV (default)

```sh
./cwexport local -from=2022-03-14T16:00:00Z -ns=authApi -name=challengesStarted -stat=Sum -dimension=ServiceName/auth-api-challengePostHandler92AD93BF-thIg6mklFAlF -dimension=ServiceType/AWS::Lambda::Function
```

### run-json

Requires: build

Run locally & outputs JSON

```sh
./cwexport local -from=2022-03-14T16:00:00Z -ns=authApi -name=challengesStarted -stat=Sum -dimension=ServiceName/auth-api-challengePostHandler92AD93BF-thIg6mklFAlF -dimension=ServiceType/AWS::Lambda::Function -format=JSON
```

### run-dynamodb-docker

```sh
docker run -p 8000:8000 amazon/dynamodb-local
```

### run-dynamodb

Run DynamoDB locally.

```sh
./run-dynamodb-local.sh
```

### test

Note: to run the tests, ensure you have a running dynamodb and you've run the build script first.

```sh
go test ./... -short
```

### test-all

```sh
go test ./...
```

### build

Build the cwexport executable.

```sh
./build.sh
```

### deploy

Requires: build

Deploy the Lambda function to the AWS environment.

```sh
./build.sh && ./cwexport deploy -config=test-config.toml
```

### get-lambda-invocations

Requires: build

```sh
./cwexport local -from=2022-03-21T16:00:00Z -ns="AWS/Lambda" -name=Invocations -stat=Sum
```
