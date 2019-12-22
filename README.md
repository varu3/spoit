# spoit
spoit is a minimal and simple deployment tool for [AWS EC2 SpotInstance](https://aws.amazon.com/jp/ec2/spot/).

spoit does, 
- Upload user-data script to S3 Bucket.
- Request EC2 spot Instance and run user-data script pulled from S3 bucket.
- output CloudWatch log groups.

## install

```
$ go get -u github.com/varusan/spoit/cmd/spoit
```

## Quick Start

```
$ mkdir hoge
$ cd hoge
$ spoit init
2019/12/11 10:52:59 [info] spoit current
2019/12/11 10:52:59 [info] creating instance.json
201

# set AWS EC2 config.
$ cat instance.json
{
  "AvailabilityZoneGroup": "",
  "SpotPrice": "",
  "LaunchSpecification": {
    "ImageId": "",
    "InstanceType": "",
    "KeyName": "",
    "SubnetId": "",
    "IamInstanceProfile": {
      "Name": ""
    },
    "SecurityGroupIds": [
      ""
    ]
  }
}

# script to execute
$ cat user-data.sh
#!/bin/bash
echo "Hello world"
```

```
# set AWS API config
$ export AWS_ACCESS_KEY_ID=""
$ export AWS_SECRET_ACCESS_KEY=""
$ export AWS_REGION="ap-northeast-1"

# dry-run
$ spoit run --bucket spoit-test --dry-run 
2019/12/11 11:08:49 [info] spoit current
2019/12/11 11:08:49 {
  AvailabilityZoneGroup: "ap-northeast-1",
  DryRun: true,
  InstanceCount: 1,
  LaunchSpecification: {
    IamInstanceProfile: {
      Name: "XXXXXXXXXXXXX"
    },
    ImageId: "XXXXXXXXXXXXX",
    InstanceType: "XXXXXXXXXXXX",
    KeyName: "XXXXXXXXXXXXX",
    SecurityGroupIds: ["XXXXXXXXXXX"],
    SubnetId: "XXXXXXXXXX",
    UserData: "IyEvYmluL2Jhc2gKY3VybCBodHRwczovL3MzLmFtYXpvbmF3cy5jb20vYXdzLWNsb3Vkd2F0Y2gvZG93bmxvYWRzL2xhdGVzdC9hd3Nsb2dzLWFnZW50LXNldHVwLnB5IC1PCnB5dGhvbiBhd3Nsb2dzLWFnZW50LXNldHVwLnB5IC1uIC1yIGFwLW5vcnRoZWFzdC0xIC1jIHMzOi8vc3BvaXQtdGVzdC9hd3Nsb2dzLmNvbmYKbWtkaXIgLXAgL3Zhci9saWIvYXdzbG9ncy8KYXdzIHMzIGNwIHMzOi8vc3BvaXQtdGVzdC91c2VyLWRhdGEuc2ggdXNlci1kYXRhLnNoCmNobW9kIHUreCB1c2VyLWRhdGEuc2gKCmVjaG8gIlJ1biBzcG9pdCB1c2VyLXNjcmlwdCIKc291cmNlIC4vdXNlci1kYXRhLnNoID4gL3RtcC9zcG9pdC1zY3JpcHQubG9nIDI+JjEKCnNsZWVwIDEwCnNodXRkb3duIC1oIG5vdwo="
  },
  SpotPrice: "0.XX"
}
2019/12/11 11:08:49 [error] failed to launch spot instance: DryRunOperation: Request would have succeeded, but DryRun flag is set.
	status code: 412, request id: d414d468-7584-4b44-af02-72d4a3517350

# Run
$ spoit run --bucket spoit-test
2019/12/11 11:15:14 [info] spoit current
2019/12/11 11:15:15 [info] upload user-data.sh to s3://spoit-test/user-data.sh
2019/12/11 11:15:15 [info] upload /var/folders/9g/f8d3ft8x6hs8z2f85hh_xvd40000gp/T/awslogs-config-tmp-914084764 to s3://spoit-test/awslogs.conf
2019/12/11 11:15:15 [info] Request spot instances... SpotInstanceRequestIDs: [sir-kterbrpg]
2019/12/11 11:15:30 [info] Launched spot instance i-00b611e0bbc244c55
2019/12/11 11:15:30 [info] output CloudWatch log group: spoit-script-logs
2019/12/11 11:15:30 [info] completed
```

then, user-data logs output cloudwath log-group.(Default: spoit-script-logs)

```
$ aws logs get-log-events --log-group-name=spoit-script-logs --log-stream-name=i-00b611e0bbc244c55
{
    "nextForwardToken": "f/35146875505599779811661892019276930066329913353797566466",
    "events": [
        {
            "ingestionTime": 1576040411651,
            "timestamp": 1576040405497,
            "message": "hello world"
        }
    ],
    "nextBackwardToken": "b/35146875505577479066463361396135394348057264992291586048"
}
```

## Launch Instance OS

Supported, 

* CentOS 7
* Amazon Linux 2

## Usage

```
usage: spoit [<flags>] <command> [<args> ...]

Flags:
  --help            Show context-sensitive help (also try --help-long and --help-man).
  --log-level=info  log level (trace, debug, info, warn, error)
  --region=""       AWS region

Commands:
  help [<command>...]
    Show help.

  version
    show version

  init
    init instance config file

  run --bucket=BUCKET [<flags>]
    run script on spot instances
```

## Load AWS credentials

spoit requires AWS regions and credentials configuration.

for regions

AWS_DEFAULT_REGION environment variable.

for credentials,

1. Environment variables.(`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
1. Shared credentials file.(Default: `~/.aws/credentials`)  
1. IAM Role fetched by Instance profile for Amazon EC2 if running on EC2 Instance.

## Run

```
usage: spoit run --bucket=BUCKET [<flags>]

run script on spot instances

Flags:
  --help                      Show context-sensitive help (also try --help-long and --help-man).
  --log-level=info            log level (trace, debug, info, warn, error)
  --region=""                 AWS region
  --instance="instance.json"  Spot instance file path
  --script="user-data.sh"     Spot instance user-data file path
  --bucket=BUCKET             S3 bucket name has user-data script
  --concurrency=1             Spot instances Concurrency number
  --log-group="spoit-script-logs"
                              CloudWatch logs group that output user-data script log
  --dry-run                   spot instnce request dry-run flag
```

### DryRun

```
spoit run --dry-run --bucket spoit-test --region ap-northeast-1
2019/12/11 15:17:52 [info] spoit current
2019/12/11 15:17:52 {
  AvailabilityZoneGroup: "ap-northeast-1a",
  DryRun: true,
  InstanceCount: 1,
  LaunchSpecification: {
    IamInstanceProfile: {
      Name: "XXXXXXXXXXXXX"
    },
    ImageId: "XXXXXXXXXXXXXX",
    InstanceType: "XXXXXXXXXX",
    KeyName: "XXXXXXXXXXX",
    SecurityGroupIds: ["XXXXXXXXXXX"],
    SubnetId: "XXXXXXXXXXXX",
    UserData: "IyEvYmluL2Jhc2gKY3VybCBodHRwczovL3MzLmFtYXpvbmF3cy5jb20vYXdzLWNsb3Vkd2F0Y2gvZG93bmxvYWRzL2xhdGVzdC9hd3Nsb2dzLWFnZW50LXNldHVwLnB5IC1PCnB5dGhvbiBhd3Nsb2dzLWFnZW50LXNldHVwLnB5IC1uIC1yIGFwLW5vcnRoZWFzdC0xIC1jIHMzOi8vc3BvaXQtdGVzdC9hd3Nsb2dzLmNvbmYKbWtkaXIgLXAgL3Zhci9saWIvYXdzbG9ncy8KYXdzIHMzIGNwIHMzOi8vc3BvaXQtdGVzdC91c2VyLWRhdGEuc2ggdXNlci1kYXRhLnNoCmNobW9kIHUreCB1c2VyLWRhdGEuc2gKCmVjaG8gIlJ1biBzcG9pdCB1c2VyLXNjcmlwdCIKc291cmNlIC4vdXNlci1kYXRhLnNoID4gL3RtcC9zcG9pdC1zY3JpcHQubG9nIDI+JjEKCnNsZWVwIDEwCnNodXRkb3duIC1oIG5vdwo="
  },
  SpotPrice: "0.04"
}
2019/12/11 15:17:52 [error] failed to launch spot instance: DryRunOperation: Request would have succeeded, but DryRun flag is set.
	status code: 412, request id: 1ba33a70-7eb2-4ae6-ac72-ee5b0a2f01eb
```

### Concurrency

Set number of concurrent executions


```
$ spoit run --bucket spoit-test --region ap-northeast-1 --concurrency 5
2019/12/11 15:19:17 [info] spoit current
2019/12/11 15:19:18 [info] upload user-data.sh to s3://spoit-test/user-data.sh
2019/12/11 15:19:18 [info] upload /var/folders/9g/f8d3ft8x6hs8z2f85hh_xvd40000gp/T/awslogs-config-tmp-317513489 to s3://spoit-test/awslogs.conf
2019/12/11 15:19:19 [info] Request spot instances... SpotInstanceRequestIDs: [sir-7gfi9vih sir-fqxraaph sir-h67ia9ag sir-jm4i8f9k sir-nv4gb6hk]
2019/12/11 15:19:34 [info] Launched spot instance i-0bb9f9eb83c568fc3
2019/12/11 15:19:34 [info] Launched spot instance i-01ded90a0b1a89d9c
2019/12/11 15:19:34 [info] Launched spot instance i-05ec803bc290b2df0
2019/12/11 15:19:34 [info] Launched spot instance i-0fb63946b568aed22
2019/12/11 15:19:34 [info] Launched spot instance i-09f1d1a394d8b1d76
2019/12/11 15:19:34 [info] output CloudWatch log group: spoit-script-logs
2019/12/11 15:19:34 [info] completed
```
