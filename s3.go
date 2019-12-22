package spoit

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

func (app *App) UploadAwslogsConfig(opt RunOption) error {

	awslogsConfTpl := `[general]
state_file = /var/lib/awslogs/agent-state

[/tmp/spoit-script.log]
datetime_format = %Y-%m-%d %H:%M:%S
file = /tmp/spoit-script.log
buffer_duration = 5000
initial_position = start_of_file
log_group_name = {{ .LogGroupName }}
log_stream_name = {instance_id}
`
	tpl, _ := template.New("").Parse(awslogsConfTpl)
	d := map[string]string{
		"LogGroupName": *opt.LogGroup,
	}
	var awslogsConf bytes.Buffer
	if err := tpl.Execute(&awslogsConf, d); err != nil {
		return errors.Wrap(err, "failed to generate awslogs config.")
	}

	tmp, err := ioutil.TempFile(os.TempDir(), "awslogs-config-tmp-")
	if err != nil {
		return errors.Wrap(err, "failed to create tmp file.")
	}
	defer os.Remove(tmp.Name())

	if err = ioutil.WriteFile(tmp.Name(), awslogsConf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "failed to write awslogs conf.")
	}

	if err = app.uploadToS3(tmp.Name(), *opt.Bucketname, "awslogs.conf"); err != nil {
		return errors.Wrap(err, "failed to upload.")
	}

	return nil
}

func (app *App) UploadUserData(opt RunOption) error {

	bucketname := *opt.Bucketname
	res, err := app.s3.ListBuckets(nil)
	if err != nil {
		return errors.Wrap(err, "failed to list s3 buckets.")
	}

	bucketCreateFlag := true
	for _, b := range res.Buckets {
		if bucketname == *b.Name {
			log.Printf("[debug] bucketname: %s is exist. skip create bucket.", bucketname)
			bucketCreateFlag = false
			break
		}
	}

	if bucketCreateFlag == true {
		log.Printf("[info] creating S3 bucket.")
		_, err := app.s3.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketname),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create s3 bucket.")
		}
	}

	if err = app.uploadToS3(ScriptFilename, bucketname, ScriptFilename); err != nil {
		return errors.Wrap(err, "failed to upload.")
	}

	return nil
}

func (app *App) uploadToS3(filename string, bucketname string, key string) error {
	log.Printf("[info] upload %s to s3://%s/%s", filename, bucketname, key)
	file, err := os.Open(filename)
	if err != nil {
		return errors.Wrap(err, "failed to read file.")
	}
	_, err = app.s3uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return errors.Wrap(err, "failed to upload file.")
	}

	return nil

}
