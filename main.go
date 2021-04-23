package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"kube-sqs-autoscaler/config"
	"kube-sqs-autoscaler/scale"
	kubesqs "kube-sqs-autoscaler/sqs"

	log "github.com/sirupsen/logrus"
)

var (
	awsRegion           string
	kubernetesNamespace string
	dryRun              bool
)

type ScalingTimeDiff struct {
	t              *time.Time
	CoolDownPeriod config.Duration
}

func (s *ScalingTimeDiff) CoolDownPassed() bool {
	if s.t == nil {
		t := time.Now()
		s.t = &t
		return false
	}

	return s.t.Add(s.CoolDownPeriod.ToDuration()).Before(time.Now())
}

func (s *ScalingTimeDiff) Reset() {
	s.t = nil
}

func Run(p *scale.PodAutoScaler, sqs *kubesqs.SqsClient, cfg *config.ScalerConfig) {
	ctx := context.Background()
	lastScalingTime := &ScalingTimeDiff{CoolDownPeriod: cfg.CoolDownPeriod}
	zeroScalingTime := &ScalingTimeDiff{CoolDownPeriod: cfg.ZeroScalingCoolDown}

	pollInterval := cfg.PollInterval.ToDuration()
	for {
		time.Sleep(pollInterval)

		numMessages, err := sqs.NumMessages()
		if err != nil {
			log.Errorf("Failed to get SQS messages: %v", err)
			continue
		}

		if numMessages == 0 && zeroScalingTime.CoolDownPassed() == false {
			log.Info("Have 0 messages but waiting for cooldown period")
			continue
		}

		if numMessages > 0 && lastScalingTime.CoolDownPassed() == false {
			log.Infof("Waiting for cooldown period to pass. current num of messages: %d", numMessages)
			continue
		}

		if err := p.Scale(ctx, numMessages); err != nil {
			log.Errorf("Failed scale up: %v", err)
			continue
		}

		lastScalingTime.Reset()
		zeroScalingTime.Reset()
	}
}

func main() {
	var configs config.ConfigFlag
	flag.Var(&configs, "config", "")
	flag.StringVar(&kubernetesNamespace, "kubernetes-namespace", "default", "The namespace your deployment is running in")
	flag.StringVar(&awsRegion, "aws-region", "", "Your AWS region")
	flag.BoolVar(&dryRun, "dry-run", true, "if scaling should run on dry-run mode or not")
	flag.Parse()

	parsedConfigs, err := config.ParseConfigFlags(configs)
	if err != nil {
		log.Errorf("Failed to parse config flags. err: %s", err)
		os.Exit(1)
	}

	for _, c := range parsedConfigs {
		// start a go routine for each tracked deployment
		go func(conf *config.ScalerConfig) {
			p := scale.NewPodAutoScaler(conf.KubernetesDeploymentName, kubernetesNamespace, conf.MaxPods, 1, conf.MessagePerPod, conf.ZeroScaling, dryRun)
			sqs := kubesqs.NewSqsClient(conf.SqsQueueUrl, awsRegion)

			log.Info(fmt.Sprintf("Starting kube-sqs-autoscaler for %s", conf.KubernetesDeploymentName))
			Run(p, sqs, conf)
		}(c)
	}

	for {
		time.Sleep(10 * time.Second)
		log.Info("health tick")
	}
}
