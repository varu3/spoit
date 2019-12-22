package spoit

import (
	"encoding/json"
)

type RunOption struct {
	Region           *string
	InstanceFilename *string
	ScriptFilename   *string
	Bucketname       *string
	Concurrency      *int64
	DryRun           *bool
	LogGroup         *string
}

func (opt *RunOption) String() string {
	b, _ := json.Marshal(opt)
	return string(b)
}

func (app *App) Run(opt RunOption) error {
	if *opt.DryRun == false {
		if err := app.UploadUserData(opt); err != nil {
			return err
		}

		if err := app.UploadAwslogsConfig(opt); err != nil {
			return err
		}
	}

	err := app.LaunchSpotInstance(opt)
	if err != nil {
		return err
	}

	return nil
}
