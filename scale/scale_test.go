package scale

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

func TestScaleUp(t *testing.T) {
	ctx := context.Background()
	p := NewMockPodAutoScaler("deploy", "namespace", 5, 1, 3)

	// Scale up replicas until we reach the max (5).
	// Scale up again and assert that we get an error back when trying to scale up replicas pass the max
	res := p.Scale(ctx, 75)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Nil(t, res.Err)
	assert.Equal(t, int32(4), *deployment.Spec.Replicas)
	res = p.Scale(ctx, 120)
	assert.Nil(t, res.Err)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(5), *deployment.Spec.Replicas)

	res = p.Scale(ctx, 250)
	assert.Nil(t, res.Err)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(5), *deployment.Spec.Replicas)
}

func TestScaleUpWithScalingPodNum(t *testing.T) {
	ctx := context.Background()
	p := NewMockPodAutoScaler("deploy", "namespace", 10, 1, 3)

	// Scale up replicas until we reach the max (10) with 5 pods scaling
	res := p.Scale(ctx, 195)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Nil(t, res.Err)
	assert.Equal(t, int32(10), *deployment.Spec.Replicas)
}

func TestScaleDown(t *testing.T) {
	ctx := context.Background()
	p := NewMockPodAutoScaler("deploy", "namespace", 5, 1, 3)

	res := p.Scale(ctx, 15)
	assert.Nil(t, res.Err)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(1), *deployment.Spec.Replicas)
}

func TestScaleDownWithScalingPodNum(t *testing.T) {
	ctx := context.Background()
	p := NewMockPodAutoScaler("deploy", "namespace", 10, 1, 8)

	res := p.Scale(ctx, 55)
	assert.Nil(t, res.Err)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(3), *deployment.Spec.Replicas)

	res = p.Scale(ctx, 10)
	assert.Nil(t, res.Err)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(1), *deployment.Spec.Replicas)
}

func TestScaleDownNotDoAnythingWhenDryRun(t *testing.T) {
	ctx := context.Background()
	p := NewMockPodAutoScaler("deploy", "namespace", 500, 1, 3)
	p.DryRun = true

	res := p.Scale(ctx, 0)
	deployment, _ := p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Nil(t, res.Err)
	assert.Equal(t, int32(3), *deployment.Spec.Replicas)

	res = p.Scale(ctx, 10000)
	assert.Nil(t, res.Err)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(3), *deployment.Spec.Replicas)

	res = p.Scale(ctx, 20)
	assert.Nil(t, res.Err)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(3), *deployment.Spec.Replicas)

	res = p.Scale(ctx, 20000)
	assert.Nil(t, res.Err)
	deployment, _ = p.Client.Get(ctx, "deploy", metav1.GetOptions{})
	assert.Equal(t, int32(3), *deployment.Spec.Replicas)
}

func NewMockPodAutoScaler(kubernetesDeploymentName string, kubernetesNamespace string, max, min, init int) *PodAutoScaler {
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
	return &PodAutoScaler{
		Client:        mock.AppsV1().Deployments(kubernetesNamespace),
		Min:           min,
		Max:           max,
		Deployment:    kubernetesDeploymentName,
		Namespace:     kubernetesNamespace,
		ZeroScaling:   false,
		MessagePerPod: 20,
	}
}
