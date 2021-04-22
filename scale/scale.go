package scale

import (
	"context"
	"os"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"math"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedappv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeConfigPath string
)

type PodAutoScaler struct {
	Client        typedappv1.DeploymentInterface
	Max           int
	Min           int
	Deployment    string
	Namespace     string
	ZeroScaling   bool
	MessagePerPod int
}

func NewPodAutoScaler(kubernetesDeploymentName string, kubernetesNamespace string, max, min, messagePerPod int, zeroScaling bool) *PodAutoScaler {
	kubeConfigPath = os.Getenv("KUBE_CONFIG_PATH")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic("Failed to configure incluster or local config")
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic("Failed to configure client")
	}

	return &PodAutoScaler{
		Client:        k8sClient.AppsV1().Deployments(kubernetesNamespace),
		Min:           min,
		Max:           max,
		Deployment:    kubernetesDeploymentName,
		Namespace:     kubernetesNamespace,
		ZeroScaling:   zeroScaling,
		MessagePerPod: messagePerPod,
	}
}

func (p *PodAutoScaler) Scale(ctx context.Context, numMessages int) error {
	deployment, err := p.Client.Get(ctx, p.Deployment, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to get deployment from kube server, no scale down occured")
	}

	currentReplicas := deployment.Spec.Replicas
	desiredReplicas := p.getDesiredReplicaCount(numMessages)

	if *currentReplicas == desiredReplicas {
		log.Infof("Same as desired replicas. Current replicas: %d Desired replicas: %d", *deployment.Spec.Replicas, desiredReplicas)
		return nil
	}

	deployment.Spec.Replicas = &desiredReplicas

	_, err = p.Client.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to scale down")
	}

	log.Infof("Scale down successful. Replicas: %d", *deployment.Spec.Replicas)
	return nil
}

func (p *PodAutoScaler) getDesiredReplicaCount(numMessages int) int32 {
	desiredReplicas := int(math.Ceil(float64(numMessages) / float64(p.MessagePerPod)))

	if p.ZeroScaling == false && desiredReplicas < int(p.Min) {
		log.Infof("desired replicas are less than min pods resetting to min. Min pod: %d Desired replicas: %d", p.Min, desiredReplicas)
		desiredReplicas = int(p.Min)
	}

	if desiredReplicas > int(p.Max) {
		log.Infof("desired replicas are more than max pods resetting to max. Max pod: %d Desired replicas: %d", p.Max, desiredReplicas)
		desiredReplicas = int(p.Max)
	}

	return int32(desiredReplicas)
}
