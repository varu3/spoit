package spoit

import (
	"bytes"
	"encoding/base64"
	"log"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func (app *App) LaunchSpotInstance(opt RunOption) error {
	log.Printf("[debug] run options: %s", opt.String())

	scriptTpl := `#!/bin/bash
set -xe -o pipefail 

export AWS_DEFAULT_REGION={{.Region}}

os=$(cat /etc/os-release | head -1 | awk '{print $3}' FS='[="]' | sed -e 's/.Linux//g')
instance_id=$(curl -s 169.254.169.254/latest/meta-data/instance-id/)
yum -y install jq
curl -kL https://bootstrap.pypa.io/get-pip.py | python 

case "$os" in 
  "CentOS" ) 
    curl {{.awslogsLatestURL}} -O
    python awslogs-agent-setup.py -n -r {{.Region}} -c s3://{{.BucketName}}/awslogs.conf
  ;;
  "Amazon" ) 
	yum -y install awslogs
	# Note: https://forums.aws.amazon.com/thread.jspa?threadID=312234
	pip install awscli==1.16.263
	sed -i s/us-east-1/{{.Region}}/g /etc/awslogs/awscli.conf
	aws s3 cp s3://{{.BucketName}}/awslogs.conf /etc/awslogs/awslogs.conf
	systemctl restart awslogsd
  ;;
esac

mkdir -p /var/lib/awslogs/
aws s3 cp s3://{{.BucketName}}/{{.ScriptFilename}} {{.ScriptFilename}}
chmod u+x {{.ScriptFilename}}

set +ex +o pipefail 
echo "Run spoit user-script"
source ./{{.ScriptFilename}} > /tmp/spoit-script.log 2>&1
echo "Done" >> /tmp/spoit-script.log 2>&1

max=10
for ((i=0; i<$max; i++)); do
  LATEST_MSG=$(aws logs get-log-events --log-group-name '{{ .LogGroupName }}' --log-stream-name $instance_id | jq -r .events[-1].message)
  if [ $LATEST_MSG = "Done" ]; then break; fi;
  sleep 10
done

shutdown -h now
`

	tpl, _ := template.New("").Parse(scriptTpl)
	d := map[string]string{
		"awslogsLatestURL": awslogsLatestURL,
		"BucketName":       *opt.Bucketname,
		"ScriptFilename":   *opt.ScriptFilename,
		"Region":           *opt.Region,
		"LogGroupName":     *opt.LogGroup,
	}

	var script bytes.Buffer
	if err := tpl.Execute(&script, d); err != nil {
		return errors.Wrap(err, "failed to generate user-data script.")
	}
	scriptStr := script.String()
	scriptEnc := base64.StdEncoding.EncodeToString([]byte(scriptStr))

	spot, err := app.loadSpotInstance(*opt.InstanceFilename)
	if err != nil {
		return errors.Wrap(err, "failed to load instance config.")
	}
	spot.LaunchSpecification.UserData = aws.String(scriptEnc)

	input := &ec2.RequestSpotInstancesInput{
		InstanceCount:         aws.Int64(*opt.Concurrency),
		DryRun:                aws.Bool(*opt.DryRun),
		AvailabilityZoneGroup: spot.AvailabilityZoneGroup,
		SpotPrice:             spot.SpotPrice,
		LaunchSpecification:   spot.LaunchSpecification,
	}

	log.Println("[debug] Requests spot instances params\n", input)
	res, err := app.ec2.RequestSpotInstances(input)
	if err != nil {
		if *opt.DryRun == true {
			log.Printf("%v", input)
		}
		return errors.Wrap(err, "failed to launch spot instance")
	}

	spRequests := []string{}
	for _, sp := range res.SpotInstanceRequests {
		spRequests = append(spRequests, *sp.SpotInstanceRequestId)
	}

	for i := 0; i < 5; i++ {
		log.Printf("[info] Request spot instances... SpotInstanceRequestIDs: %v", spRequests)
		err = app.ec2.WaitUntilSpotInstanceRequestFulfilled(&ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: aws.StringSlice(spRequests),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidSpotInstanceRequestID.NotFound" {
					continue
				}
			}
			log.Println("[error] Failed fulfilling spot request")
		}
		break
	}

	resp, err := app.ec2.DescribeSpotInstanceRequests(&ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: aws.StringSlice(spRequests),
	})
	if err != nil {
		log.Printf("[error] Failed to describe spot instance")
	}

	log.Printf("[debug] Launched spot instance params:\n%v", resp)
	for _, inst := range resp.SpotInstanceRequests {
		log.Printf("[info] Launched spot instance %v", *inst.InstanceId)
	}

	log.Printf("[info] output CloudWatch log group: %s", *opt.LogGroup)
	return nil
}
