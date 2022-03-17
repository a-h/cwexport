package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/a-h/cwexport/deploycmd"
	"github.com/a-h/cwexport/localcmd"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// Binary builds set this version string. goreleaser sets the value using Go build ldflags.
var version string

// Source builds use this value. When installed using `go install github.com/a-h/templ/cmd/templ@latest` the `version` variable is empty, but
// the debug.ReadBuildInfo return value provides the package version number installed by `go install`
func goInstallVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}

func getVersion() string {
	if version != "" {
		return version
	}
	return goInstallVersion()
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "local":
		localCmd(os.Args[2:])
		return
	case "deploy":
		deployCmd(os.Args[2:])
		return
	case "version":
		fmt.Println(getVersion())
		return
	case "--version":
		fmt.Println(getVersion())
		return
	}
	usage()
}

func usage() {
	fmt.Println(`usage: cwexport <command> [parameters]
To see help text, you can run:
  cwexport local --help
  cwexport deploy --help
  cwexport version
examples:
  cwexport local -from=2022-03-14T16:00:00Z -ns=authApi -name=challengesStarted -stat=Sum -dimension=ServiceName/auth-api-challengePostHandler92AD93BF-thIg6mklFAlF -dimension=ServiceType/AWS::Lambda::Function
  cwexport deploy`)
	os.Exit(1)
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func localCmd(args []string) {
	cmd := flag.NewFlagSet("local", flag.ExitOnError)
	from := cmd.String("from", "", "The time to start exporting.")
	namespace := cmd.String("ns", "", "The namespace of the metric.")
	name := cmd.String("name", "", "The name of the metric.")
	stat := cmd.String("stat", "Sum", "The stat to use, e.g. Sum or Average.")
	var dimensions arrayFlags
	cmd.Var(&dimensions, "dimension", "Dimension as key value, e.g. ServiceName/123")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}

	var messages []string
	var cmdArgs localcmd.Args

	if cmdArgs.Start, err = time.Parse(time.RFC3339, *from); err != nil {
		messages = append(messages, "Missing or invalid 'from' date parameter")

	}
	if *namespace == "" {
		messages = append(messages, "Missing 'ns' string parameter")
	}
	if *name == "" {
		messages = append(messages, "Missing 'name' string parameter")
	}
	dims := make([]types.Dimension, len(dimensions))
	for i := 0; i < len(dimensions); i++ {
		v := strings.SplitN(dimensions[i], "/", 2)
		if len(v) != 2 {
			messages = append(messages, fmt.Sprintf("Invalid dimension %q", dimensions[i]))
			continue
		}
		dims[i] = types.Dimension{
			Name:  &v[0],
			Value: &v[1],
		}
	}
	if len(messages) > 0 {
		fmt.Println("Errors:")
		for _, m := range messages {
			fmt.Printf("  %s\n", m)
		}
	}

	cmdArgs.MetricStat = &types.MetricStat{
		Metric: &types.Metric{
			Dimensions: dims,
			MetricName: name,
			Namespace:  namespace,
		},
		Period: aws.Int32(int32((5 * time.Minute).Seconds())),
		Stat:   stat,
	}

	err = localcmd.Run(cmdArgs)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
  }
}

func deployCmd(args []string) {
	cmd := flag.NewFlagSet("deploy", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	configFlag := cmd.String("config", "", "Config file")

	var messages []string

	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}

  if *configFlag == "" {
		messages = append(messages, "Missing config file")
	}

	confData, err := ioutil.ReadFile(*configFlag)
	if err != nil {
		messages = append(messages, "Unable to read config")
	}

	var ms types.MetricStat
	_, err = toml.Decode(string(confData), &ms)
	if err != nil {
		messages = append(messages, "Unable to parse config file")
	}

	if len(messages) > 0 {
		fmt.Println("Errors:")
		for _, m := range messages {
			fmt.Printf("  %s\n", m)
		}
		os.Exit(1)
	}

	err = deploycmd.Run(&ms)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
