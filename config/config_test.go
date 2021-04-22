package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	f := &ConfigFlag{}
	f.Set(`{
		"pollInterval": "5s" ,
		"coolDownPeriod": "300s",
		"messagePerPod": 100,
		"maxPods": 10,
		"zeroScaling": false,
		"zeroScalingCoolDown": "300s",
		"sqsQueueUrl": "some-queue-url",
		"deploymentName": "deployment-name"
	 }`)
	cfgs, err := ParseConfigFlags(*f)
	assert.Nil(t, err)
	for _, c := range cfgs {
		assert.Equal(t, c.CoolDownPeriod, 300*time.Second)
		assert.Equal(t, c.PollInterval, 5*time.Second)
		assert.Equal(t, c.MessagePerPod, 100)
		assert.Equal(t, c.MaxPods, 10)
		assert.Equal(t, c.ZeroScaling, false)
		assert.Equal(t, c.ZeroScalingCoolDown, 300*time.Second)
		assert.Equal(t, c.SqsQueueUrl, "some-queue-url")
		assert.Equal(t, c.KubernetesDeploymentName, "deployment-name")
	}
}

func TestParseFailureInvalidJson(t *testing.T) {
	tests := []string{"aslkcnasc", "xv df,,sd", "123"}

	for _, tt := range tests {
		f := &ConfigFlag{}
		f.Set(tt)
		_, err := ParseConfigFlags(*f)
		assert.NotNil(t, err, fmt.Sprintf("val: %s", tt))
	}

}

func TestParseFailureDataValidation(t *testing.T) {
	tests := []string{
		"{}",
		`{
			"pollInterval": "5s" ,
			"coolDownPeriod": "300s",
			"messagePerPod": 0,
			"maxPods": 10,
			"zeroScaling": false,
			"zeroScalingCoolDown": "300s",
			"sqsQueueUrl": "some-queue-url",
			"deploymentName": "deployment-name"
		 }`,
		`{
			"pollInterval": "5s" ,
			"coolDownPeriod": "300s",
			"messagePerPod": 100,
			"maxPods": 0,
			"zeroScaling": false,
			"zeroScalingCoolDown": "300s",
			"sqsQueueUrl": "some-queue-url",
			"deploymentName": "deployment-name"
		 }`,
		`{
			"pollInterval": "5s" ,
			"coolDownPeriod": "300s",
			"messagePerPod": 100,
			"maxPods": 10,
			"zeroScaling": false,
			"zeroScalingCoolDown": "300s",
			"sqsQueueUrl": "",
			"deploymentName": "deployment-name"
		 }`,
		`{
			"pollInterval": "5s" ,
			"coolDownPeriod": "300s",
			"messagePerPod": 100,
			"maxPods": 10,
			"zeroScaling": false,
			"zeroScalingCoolDown": "300s",
			"sqsQueueUrl": "some-queue-url",
			"deploymentName": ""
		 }`,
	}

	for _, tt := range tests {
		f := &ConfigFlag{}
		f.Set(tt)
		_, err := ParseConfigFlags(*f)
		assert.NotNil(t, err)
	}

}
