# How to test with minikube
* Start minikube
`minikube start`
* apply 
`kubectl apply -f minikube/`
* tail logs from autoscaler pod
`kubectl logs -f kube-sqs-autoscaler-56c5c4cbdb-pgs9d`
* expose elstic mq
`minikube service elasticmq`
* generate messages for queues
`python generate_messages.py test-queue1 PORT_FROM_SERVICE 35`
`python generate_messages.py test-queue2 PORT_FROM_SERVICE 85`
* purge queue or read 1 message per second with the python script
`aws sqs --region eu-central-1 --endpoint-url http://127.0.0.1:PORT_FROM_SERVICE purge-queue --queue-url http://localhost:9324/queue/test-queue1`
`aws sqs --region eu-central-1 --endpoint-url http://127.0.0.1:PORT_FROM_SERVICE purge-queue --queue-url http://localhost:9324/queue/test-queue2`