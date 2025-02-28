.PHONY: test clean compile build push

IMAGE=evardar/kube-sqs-autoscaler
VERSION=v1.0.0

test:
	go test -race ./...

clean:
	rm -f kube-sqs-autoscaler

compile: clean
	GOOS=linux go build .

build: compile
	docker build -t $(IMAGE):$(VERSION) .

push: build
	docker push $(IMAGE):$(VERSION)
