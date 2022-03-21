# cwexport

Exports CloudWatch metrics.

## Tasks

### run

Run locally & print out metrics as CSV (TODO: support for JSON output)
```sh
./build.sh && ./cwexport local \
    -from=2022-03-14T16:00:00Z \
    -ns=authApi \
    -name=challengesStarted \
    -stat=Sum \
    -dimension=ServiceName/auth-api-challengePostHandler92AD93BF-thIg6mklFAlF \
    -dimension=ServiceType/AWS::Lambda::Function
```

For more information see:
```sh
./build.sh && ./cwexport local --help
```

### run-dynamodb

* Using docker:
```sh
docker run -p 8000:8000 amazon/dynamodb-local
```
* Using locally running Dynamodb instance
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
### Deploy

```sh
./build.sh && ./cwexport deploy -config=test-config.toml
```

For more information see:
```sh
./build.sh && ./cwexport deploy --help
```
