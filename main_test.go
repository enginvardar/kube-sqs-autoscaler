package main

import (
	"context"
	"kube-sqs-autoscaler/config"
	"kube-sqs-autoscaler/scale"
	kubesqs "kube-sqs-autoscaler/sqs"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

func TestRunScaleUpCoolDown(t *testing.T) {
	ctx := context.Background()
	awsRegion = "us-east-1"
	kubernetesNamespace = "namespace"
	initPods := 3
	p := NewMockPodAutoScaler("deploy", kubernetesNamespace, 100, 1, initPods)
	s := NewMockSqsClient()
	c := NewScalerConfig(1*time.Second, 5*time.Second, 20, 100, false, 20*time.Second, "example.com", "deploy")
	go Run(p, s, c)

	Attributes := map[string]*string{
		"ApproximateNumberOfMessages":           aws.String("80"),
		"ApproximateNumberOfMessagesDelayed":    aws.String("0"),
		"ApproximateNumberOfMessagesNotVisible": aws.String("20"),
	}

	input := &sqs.SetQueueAttributesInput{
		Attributes: Attributes,
	}
	_, _ = s.Client.SetQueueAttributes(input)

	time.Sleep(3 * time.Second)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(3), *deployment.Spec.Replicas, "Number of replicas should be 3 before the cool down period")

	time.Sleep(5 * time.Second)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(5), *deployment.Spec.Replicas, "Number of replicas should be 5 after the cool down period")
}

func TestRunScaleDownCoolDown(t *testing.T) {
	ctx := context.Background()

	awsRegion = "us-east-1"
	kubernetesNamespace = "namespace"
	initPods := 3
	p := NewMockPodAutoScaler("deploy", kubernetesNamespace, 100, 1, initPods)
	s := NewMockSqsClient()
	c := NewScalerConfig(1*time.Second, 5*time.Second, 20, 100, false, 20*time.Second, "example.com", "deploy")
	go Run(p, s, c)

	Attributes := map[string]*string{
		"ApproximateNumberOfMessages":           aws.String("1"),
		"ApproximateNumberOfMessagesDelayed":    aws.String("1"),
		"ApproximateNumberOfMessagesNotVisible": aws.String("1"),
	}

	input := &sqs.SetQueueAttributesInput{
		Attributes: Attributes,
	}
	_, _ = s.Client.SetQueueAttributes(input)

	time.Sleep(3 * time.Second)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(3), *deployment.Spec.Replicas, "Number of replicas should be 3 before the cool down period")

	time.Sleep(5 * time.Second)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(1), *deployment.Spec.Replicas, "Number of replicas should be 1 after the cool down period")
}

func TestRunReachOneReplicaWithScaleing(t *testing.T) {
	ctx := context.Background()
	awsRegion = "us-east-1"
	kubernetesNamespace = "namespace"
	initPods := 3
	p := NewMockPodAutoScaler("deploy", kubernetesNamespace, 100, 1, initPods)
	s := NewMockSqsClient()
	c := NewScalerConfig(1*time.Second, 1*time.Second, 20, 100, false, 20*time.Second, "example.com", "deploy")
	go Run(p, s, c)

	Attributes := map[string]*string{
		"ApproximateNumberOfMessages":           aws.String("1"),
		"ApproximateNumberOfMessagesDelayed":    aws.String("1"),
		"ApproximateNumberOfMessagesNotVisible": aws.String("1"),
	}

	input := &sqs.SetQueueAttributesInput{
		Attributes: Attributes,
	}
	_, _ = s.Client.SetQueueAttributes(input)

	time.Sleep(3 * time.Second)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(1), *deployment.Spec.Replicas, "Number of replicas should be the min")
}

func TestRunReachMaxReplicasWithScaleing(t *testing.T) {
	ctx := context.Background()

	awsRegion = "us-east-1"
	kubernetesNamespace = "namespace"
	initPods := 3
	p := NewMockPodAutoScaler("deploy", kubernetesNamespace, 100, 1, initPods)
	s := NewMockSqsClient()
	c := NewScalerConfig(1*time.Second, 1*time.Second, 20, 100, false, 20*time.Second, "example.com", "deploy")
	go Run(p, s, c)

	Attributes := map[string]*string{
		"ApproximateNumberOfMessages":           aws.String("5000"),
		"ApproximateNumberOfMessagesDelayed":    aws.String("5000"),
		"ApproximateNumberOfMessagesNotVisible": aws.String("5000"),
	}

	input := &sqs.SetQueueAttributesInput{
		Attributes: Attributes,
	}
	_, _ = s.Client.SetQueueAttributes(input)

	time.Sleep(3 * time.Second)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(100), *deployment.Spec.Replicas, "Number of replicas should be the max")
}

func NewMockPodAutoScaler(kubernetesDeploymentName string, kubernetesNamespace string, max, min, init int) *scale.PodAutoScaler {
	initialReplicas := int32(init)
	mock := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "deploy",
			Namespace:   "namespace",
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &initialReplicas,
		},
	}, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "deploy-no-scale",
			Namespace:   "namespace",
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &initialReplicas,
		},
	})
	return &scale.PodAutoScaler{
		Client:        mock.AppsV1().Deployments(kubernetesNamespace),
		Min:           min,
		Max:           max,
		Deployment:    kubernetesDeploymentName,
		Namespace:     kubernetesNamespace,
		ZeroScaling:   false,
		MessagePerPod: 20,
	}
}

type MockSQS struct {
	QueueAttributes *sqs.GetQueueAttributesOutput
}

func (m *MockSQS) GetQueueAttributes(*sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	return m.QueueAttributes, nil
}

func (m *MockSQS) SetQueueAttributes(input *sqs.SetQueueAttributesInput) (*sqs.SetQueueAttributesOutput, error) {
	m.QueueAttributes = &sqs.GetQueueAttributesOutput{
		Attributes: input.Attributes,
	}
	return &sqs.SetQueueAttributesOutput{}, nil
}

func NewMockSqsClient() *kubesqs.SqsClient {
	Attributes := map[string]*string{
		"ApproximateNumberOfMessages":           aws.String("100"),
		"ApproximateNumberOfMessagesDelayed":    aws.String("100"),
		"ApproximateNumberOfMessagesNotVisible": aws.String("100"),
	}

	return &kubesqs.SqsClient{
		Client: &MockSQS{
			QueueAttributes: &sqs.GetQueueAttributesOutput{
				Attributes: Attributes,
			},
		},
		QueueUrl: "example.com",
	}
}

func NewScalerConfig(pollInterval time.Duration, coolDownPeriod time.Duration, messagePerPod int, maxPods int, zeroscaling bool, zeroScalingCoolDown time.Duration, queueUrl string, deploymentName string) *config.ScalerConfig {
	p := config.Duration(pollInterval)
	cdp := config.Duration(coolDownPeriod)
	zsp := config.Duration(zeroScalingCoolDown)
	return &config.ScalerConfig{
		PollInterval:             p,
		CoolDownPeriod:           cdp,
		MessagePerPod:            messagePerPod,
		MaxPods:                  maxPods,
		ZeroScaling:              zeroscaling,
		ZeroScalingCoolDown:      zsp,
		SqsQueueUrl:              queueUrl,
		KubernetesDeploymentName: deploymentName,
	}
}
