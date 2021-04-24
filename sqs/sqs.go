package sqs

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/pkg/errors"
)

type SQS interface {
	GetQueueAttributes(*sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error)
	GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
}

type SqsClient struct {
	Client    SQS
	QueueUrl  string
	QueueName string
}

func NewSqsClient(queue string, region string) *SqsClient {
	svc := sqs.New(session.Must(session.NewSession()), aws.NewConfig().WithRegion(region))
	return &SqsClient{
		Client:    svc,
		QueueName: queue,
		QueueUrl:  "",
	}
}

func (s *SqsClient) NumMessages() (int, error) {
	if s.QueueUrl == "" {
		queuUrlInput := sqs.GetQueueUrlInput{QueueName: &s.QueueName}
		queueUrl, err := s.Client.GetQueueUrl(&queuUrlInput)
		if err != nil {
			return -1, errors.Errorf("Could not fetch queue url %s", err)
		}
		s.QueueUrl = queueUrl.String()
	}

	params := sqs.GetQueueAttributesInput{
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
			aws.String("ApproximateNumberOfMessagesDelayed"),
			aws.String("ApproximateNumberOfMessagesNotVisible")},
		QueueUrl: aws.String(s.QueueUrl),
	}

	out, err := s.Client.GetQueueAttributes(&params)
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get messages in SQS")
	}

	approximateNumberOfMessages, err := strconv.Atoi(*out.Attributes["ApproximateNumberOfMessages"])
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get number of messages in queue")
	}

	approximateNumberOfMessagesDelayed, err := strconv.Atoi(*out.Attributes["ApproximateNumberOfMessagesDelayed"])
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get number of messages in queue")
	}

	approximateNumberOfMessagesNotVisible, err := strconv.Atoi(*out.Attributes["ApproximateNumberOfMessagesNotVisible"])
	if err != nil {
		return -1, errors.Wrap(err, "Failed to get number of messages in queue")
	}

	messages := approximateNumberOfMessages + approximateNumberOfMessagesDelayed + approximateNumberOfMessagesNotVisible

	return messages, nil
}
