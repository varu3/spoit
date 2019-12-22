package spoit

import (
	"encoding/json"
	"log"
)

type SpotRequestConfig struct {
	AvailabilityZoneGroup string
	SpotPrice             string
	LaunchSpecification   *SpotInstanceRequestConfig
}

type SpotInstanceRequestConfig struct {
	ImageId            string
	InstanceType       string
	KeyName            string
	SubnetId           string
	IamInstanceProfile *IamInstanceProfile
	SecurityGroupIds   []string
}

type IamInstanceProfile struct {
	Name string
}

func (app *App) Init() error {
	js, _ := json.MarshalIndent(&SpotRequestConfig{
		AvailabilityZoneGroup: "",
		SpotPrice:             "",
		LaunchSpecification: &SpotInstanceRequestConfig{
			ImageId:      "",
			InstanceType: "",
			KeyName:      "",
			SubnetId:     "",
			IamInstanceProfile: &IamInstanceProfile{
				Name: "",
			},
			SecurityGroupIds: []string{""},
		},
	}, "", "  ")

	log.Printf("[info] creating instance.json")
	app.saveFile("./instance.json", js, 0644)

	return nil
}
