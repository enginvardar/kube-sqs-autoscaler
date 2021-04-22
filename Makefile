.PHONY: test clean compile build push

IMAGE=enginvardar/kube-sqs-autoscaler
VERSION=v1.0.0

test:
	go test ./...

clean:
	rm -f kube-sqs-autoscaler

compile: clean
	GOOS=linux go build .

build: compile
	docker build -t $(IMAGE):$(VERSION) .

push: build
	docker push $(IMAGE):$(VERSION)
