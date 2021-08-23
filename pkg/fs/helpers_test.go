package fs

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type mockS3Client struct {
	topicConfigurations                          []*s3.TopicConfiguration
	mockGetBucketNotificationConfigurationOutput *s3.NotificationConfiguration
	mockGetBucketNotificationConfigurationError  error
	mockPutBucketNotificationConfigurationError  error
}

func (m *mockS3Client) GetBucketNotificationConfiguration(input *s3.GetBucketNotificationConfigurationRequest) (*s3.NotificationConfiguration, error) {
	return m.mockGetBucketNotificationConfigurationOutput, m.mockGetBucketNotificationConfigurationError
}

func (m *mockS3Client) PutBucketNotificationConfiguration(input *s3.PutBucketNotificationConfigurationInput) (*s3.PutBucketNotificationConfigurationOutput, error) {
	m.topicConfigurations = input.NotificationConfiguration.TopicConfigurations

	return nil, m.mockPutBucketNotificationConfigurationError
}

func Test_addOrRemoveNotification(t *testing.T) {
	tests := []struct {
		description                                  string
		mockGetBucketNotificationConfigurationOutput *s3.NotificationConfiguration
		mockGetBucketNotificationConfigurationError  error
		mockPutBucketNotificationConfigurationError  error
		add                                          bool
		topicConfigurations                          []*s3.TopicConfiguration
		error                                        string
	}{
		{
			description: "error getting bucket notification",
			mockGetBucketNotificationConfigurationOutput: nil,
			mockGetBucketNotificationConfigurationError:  errors.New("mock get bucket notification configuration error"),
			mockPutBucketNotificationConfigurationError:  nil,
			add:                 false,
			topicConfigurations: nil,
			error:               "mock get bucket notification configuration error",
		},
		{
			description: "error putting bucket notification",
			mockGetBucketNotificationConfigurationOutput: &s3.NotificationConfiguration{
				TopicConfigurations: []*s3.TopicConfiguration{},
			},
			mockGetBucketNotificationConfigurationError: nil,
			mockPutBucketNotificationConfigurationError: errors.New("mock put bucket notification configuration error"),
			add:                 true,
			topicConfigurations: nil,
			error:               "mock put bucket notification configuration error",
		},
		{
			description: "successful add notification invocation",
			mockGetBucketNotificationConfigurationOutput: &s3.NotificationConfiguration{
				TopicConfigurations: []*s3.TopicConfiguration{
					{
						TopicArn: aws.String("old_topic_arn"),
					},
				},
			},
			mockGetBucketNotificationConfigurationError: nil,
			mockPutBucketNotificationConfigurationError: nil,
			add: true,
			topicConfigurations: []*s3.TopicConfiguration{
				{
					TopicArn: aws.String("old_topic_arn"),
				},
				{
					TopicArn: aws.String("new_topic_arn"),
				},
			},
			error: "",
		},
		{
			description: "successful remove notification invocation",
			mockGetBucketNotificationConfigurationOutput: &s3.NotificationConfiguration{
				TopicConfigurations: []*s3.TopicConfiguration{
					{
						TopicArn: aws.String("old_topic_arn"),
					},
					{
						TopicArn: aws.String("new_topic_arn"),
					},
				},
			},
			mockGetBucketNotificationConfigurationError: nil,
			mockPutBucketNotificationConfigurationError: nil,
			add: true,
			topicConfigurations: []*s3.TopicConfiguration{
				{
					TopicArn: aws.String("old_topic_arn"),
				},
			},
			error: "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				topicARN: "new_topic_arn",
				s3Client: &mockS3Client{
					mockGetBucketNotificationConfigurationOutput: test.mockGetBucketNotificationConfigurationOutput,
					mockGetBucketNotificationConfigurationError:  test.mockGetBucketNotificationConfigurationError,
					mockPutBucketNotificationConfigurationError:  test.mockPutBucketNotificationConfigurationError,
				},
			}

			err := h.addOrRemoveNotification(context.Background(), "bucket", test.add)

			if err != nil {
				if err.Error() != test.error {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error)
				}
			} else {
				receivedTopicConfiguration := h.s3Client.(*mockS3Client).topicConfigurations
				for i, topicConfiguration := range test.topicConfigurations {
					if *topicConfiguration.TopicArn != *receivedTopicConfiguration[i].TopicArn {
						t.Errorf("incorrect topic arn, received: %s, expected: %s", *topicConfiguration.TopicArn, *receivedTopicConfiguration[i].TopicArn)
					}
				}
			}
		})
	}
}
