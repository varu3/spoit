package main

import (
	"fmt"
	"log"
	"os"

	"github.com/varusan/spoit"
	"github.com/hashicorp/logutils"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Version Number
var (
	Version = "current"
	Region  = os.Getenv("AWS_REGION")
)

func main() {
	os.Exit(_main())
}

func _main() int {
	kingpin.Command("version", "show version")
	kingpin.Command("init", "init instance config file")
	logLevel := kingpin.Flag("log-level", "log level (trace, debug, info, warn, error)").Default("info").Enum("trace", "debug", "info", "warn", "error")
	region := kingpin.Flag("region", "AWS region").Default(Region).String()

	run := kingpin.Command("run", "run script on spot instances")
	runOption := spoit.RunOption{
		InstanceFilename: run.Flag("instance", "Spot instance file path").Default(spoit.InstanceFilename).String(),
		ScriptFilename:   run.Flag("script", "Spot instance user-data file path").Default(spoit.ScriptFilename).String(),
		Bucketname:       run.Flag("bucket", "S3 bucket name has user-data script").Required().String(),
		Concurrency:      run.Flag("concurrency", "Spot instances Concurrency number").Default(spoit.Concurrency).Int64(),
		LogGroup:         run.Flag("log-group", "CloudWatch logs group that output user-data script log").Default("spoit-script-logs").String(),
		DryRun:           run.Flag("dry-run", "spot instnce request dry-run flag").Bool(),
	}
	command := kingpin.Parse()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"trace", "debug", "info", "warn", "error"},
		MinLevel: logutils.LogLevel(*logLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	app, err := spoit.New(*region)
	if err != nil {
		log.Println("[error]", err)
		return 1
	}

	if command == "version" {
		fmt.Println("spoit", Version)
		return 0
	}

	log.Println("[info] spoit", Version)
	switch command {
	case "init":
		err = app.Init()
	case "run":
		runOption.Region = region
		err = app.Run(runOption)
	}

	if err != nil {
		log.Println("[error]", err)
		return 1
	}

	log.Println("[info] completed")
	return 0
}
