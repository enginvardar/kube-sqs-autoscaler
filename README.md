# kube-sqs-autoscaler

Kubernetes pod autoscaler based on queue size in AWS SQS. It periodically retrieves the number of messages in your queue and scales pods accordingly.

Forked https://github.com/Wattpad/kube-sqs-autoscaler

## Setting up

Setting up kube-sqs-autoscaler requires two steps:

1) Deploying it as an incluster service in your cluster
2) Adding AWS permissions so it can read the number of messages in your queues.

### Deploying kube-sqs-autoscaler

Deployin kube-sqs-autoscaler should be as simple as applying this deployment:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-sqs-autoscaler
  labels:
    app: kube-sqs-autoscaler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kube-sqs-autoscaler
  template:
    metadata:
      labels:
        app: kube-sqs-autoscaler
    spec:
      containers:
      - name: kube-sqs-autoscaler
        image: irotoris/kube-sqs-autoscaler:v2.1.0
        command:
          - /kube-sqs-autoscaler
          - --config='{"pollInterval": "5s", "coolDownPeriod": "300s", "messagePerPod": 100, "maxPods": 10, "zeroScaling": false, "zeroScalingCoolDown": "300s", "sqsQueueUrl": "https://sqs.your_aws_region.amazonaws.com/your_aws_account_number/your_queue_name", "deploymentName": "your-kubernetes-deployment-name" }'
          - --kubernetes-namespace=$(POD_NAMESPACE) # required
          - --aws-region=us-west-1  #required
        env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
        resources:
          requests:
            memory: "200Mi"
            cpu: "100m"
          limits:
            memory: "200Mi"
            cpu: "100m"
```

### Permissions

Next you want to attach this policy so kube-sqs-autoscaler can retreive SQS attributes:

```json
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "sqs:GetQueueAttributes",
        "Resource": "arn:aws:sqs:your_aws_account_number:your_region:your_sqs_queue"
    }]
}
```
