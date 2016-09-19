package scale

import (
	"github.com/pkg/errors"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

type KubeClient interface {
	Deployments(namespace string) kclient.DeploymentInterface
}

type PodAutoScaler struct {
	Client     KubeClient
	Max        int
	Min        int
	Deployment string
	Namespace  string
}

func NewPodAutoScaler(kubernetesDeploymentName string, kubernetesNamespace string, max int, min int) *PodAutoScaler {
	config, err := restclient.InClusterConfig()
	if err != nil {
		panic("Failed to configure incluster config")
	}

	k8sClient, err := kclient.New(config)
	if err != nil {
		panic("Failed to configure client")
	}

	return &PodAutoScaler{
		Client:     k8sClient,
		Min:        min,
		Max:        max,
		Deployment: kubernetesDeploymentName,
		Namespace:  kubernetesNamespace,
	}
}

func (p *PodAutoScaler) ScaleUp() error {
	deployment, err := p.Client.Deployments(p.Namespace).Get(p.Deployment)
	if err != nil {
		return errors.Wrap(err, "Failed to get deployment from kube server, no scale up occured")
	}

	currentReplicas := deployment.Spec.Replicas

	if currentReplicas >= int32(p.Max) {
		return errors.New("Max pods reached, not scaling up")
	}

	deployment.Spec.Replicas = currentReplicas + 1

	_, err = p.Client.Deployments(p.Namespace).Update(deployment)
	if err != nil {
		return errors.Wrap(err, "Failed to scale up")
	}

	return nil
}

func (p *PodAutoScaler) ScaleDown() error {
	deployment, err := p.Client.Deployments(api.NamespaceDefault).Get(p.Deployment)
	if err != nil {
		return errors.Wrap(err, "Failed to get deployment from kube server, no scale up occured")
	}

	currentReplicas := deployment.Spec.Replicas

	if currentReplicas <= int32(p.Min) {
		return errors.New("Min pods reached, not scaling down")
	}

	deployment.Spec.Replicas = currentReplicas - 1

	// TODO: use same namespace that kube-sqs-autoscaler is using
	_, err = p.Client.Deployments(api.NamespaceDefault).Update(deployment)
	if err != nil {
		return errors.Wrap(err, "Failed to scale down")
	}

	return nil
}
