package config

import (
	"encoding/json"
	"errors"
	"time"
)

type ConfigFlag []string

func (i *ConfigFlag) String() string {
	return "{}"
}

func (i *ConfigFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// https://stackoverflow.com/a/54571600
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) ToDuration() time.Duration {
	dur := time.Duration(*d)
	return dur
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type ScalerConfig struct {
	PollInterval             Duration `json:"pollInterval"`
	CoolDownPeriod           Duration `json:"coolDownPeriod"`
	MessagePerPod            int      `json:"messagePerPod"`
	MaxPods                  int      `json:"maxPods"`
	ZeroScaling              bool     `json:"zeroScaling"`
	ZeroScalingCoolDown      Duration `json:"zeroScalingCoolDown"`
	SqsQueueUrl              string   `json:"sqsQueueUrl"`
	KubernetesDeploymentName string   `json:"deploymentName"`
}

type ScalerConfigs []*ScalerConfig

func ParseConfigFlags(c ConfigFlag) (ScalerConfigs, error) {
	parsedConfigs := ScalerConfigs{}
	var err error
	for _, conf := range c {
		var sc ScalerConfig
		err = json.Unmarshal([]byte(conf), &sc)
		if err != nil {
			// bail
			return parsedConfigs, err
		}

		if isConfigValid(sc) == false {
			return parsedConfigs, errors.New("Some fields are missing")
		}
	}
	return parsedConfigs, nil
}

func isConfigValid(s ScalerConfig) bool {
	if s.MessagePerPod > 0 &&
		s.MaxPods > 1 &&
		s.SqsQueueUrl != "" &&
		s.KubernetesDeploymentName != "" {
		return true
	}
	return false
}
