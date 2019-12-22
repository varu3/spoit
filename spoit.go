package spoit

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/kayac/go-config"
	"github.com/pkg/errors"
)

type Spotrequest = ec2.RequestSpotInstancesInput

type App struct {
	sess       *session.Session
	ec2        *ec2.EC2
	s3         *s3.S3
	s3uploader *s3manager.Uploader
	accountID  string
}

func (app *App) IamFleetRoleArn(name string) string {
	return fmt.Sprintf(
		"arn:aws:iam::%s:role/%s",
		app.AWSAccountID(),
		name,
	)
}

var (
	InstanceFilename = "instance.json"
	ScriptFilename   = "user-data.sh"
	Concurrency      = "1"
	awslogsLatestURL = "https://s3.amazonaws.com/aws-cloudwatch/downloads/latest/awslogs-agent-setup.py"
)

func New(region string) (*App, error) {
	conf := &aws.Config{}
	if region != "" {
		conf.Region = aws.String(region)
	}
	sess := session.Must(session.NewSession(conf))
	return &App{
		sess:       sess,
		ec2:        ec2.New(sess),
		s3:         s3.New(sess),
		s3uploader: s3manager.NewUploader(sess),
	}, nil
}

// AWSAccountID returns AWS account ID in current session
func (app *App) AWSAccountID() string {
	if app.accountID != "" {
		return app.accountID
	}
	svc := sts.New(app.sess)
	r, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Println("[warn] failed to get caller identity", err)
		return ""
	}
	app.accountID = *r.Account
	return app.accountID
}

func (app *App) loadSpotInstance(path string) (*Spotrequest, error) {
	var spot Spotrequest

	err := config.LoadWithEnvJSON(&spot, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load %s", path)
	}

	return &spot, nil
}
