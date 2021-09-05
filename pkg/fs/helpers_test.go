package fs

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
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

type mockSNSClient struct {
	attributeValue               *string
	mockGetTopicAttributesOutput *sns.GetTopicAttributesOutput
	mockGetTopicAttributesError  error
	mockSetTopicAttributesError  error
}

func (m *mockSNSClient) GetTopicAttributes(input *sns.GetTopicAttributesInput) (*sns.GetTopicAttributesOutput, error) {
	return m.mockGetTopicAttributesOutput, m.mockGetTopicAttributesError
}

func (m *mockSNSClient) SetTopicAttributes(input *sns.SetTopicAttributesInput) (*sns.SetTopicAttributesOutput, error) {
	m.attributeValue = input.AttributeValue

	return nil, m.mockSetTopicAttributesError
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
			add: false,
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

func Test_addOrRemoveTopicPolicyBucketARN(t *testing.T) {
	tests := []struct {
		description                  string
		mockGetTopicAttributesOutput *sns.GetTopicAttributesOutput
		mockGetTopicAttributesError  error
		mockSetTopicAttributesError  error
		add                          bool
		attributeValue               *string
		error                        string
	}{
		{
			description:                  "error getting topic attributes",
			mockGetTopicAttributesOutput: nil,
			mockGetTopicAttributesError:  errors.New("mock get topic attributes error"),
			mockSetTopicAttributesError:  nil,
			add:                          true,
			attributeValue:               nil,
			error:                        "mock get topic attributes error",
		},
		{
			description: "error unmarshalling sns topic policy",
			mockGetTopicAttributesOutput: &sns.GetTopicAttributesOutput{
				Attributes: map[string]*string{
					"Policy": aws.String("---------"),
				},
			},
			mockGetTopicAttributesError: nil,
			mockSetTopicAttributesError: nil,
			add:                         true,
			attributeValue:              nil,
			error:                       "invalid character '-' in numeric literal",
		},
		{
			description: "error setting topic attributes",
			mockGetTopicAttributesOutput: &sns.GetTopicAttributesOutput{
				Attributes: map[string]*string{
					"Policy": aws.String("{\"Version\":\"2008-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"s3.amazonaws.com\"},\"Action\":\"sns:Publish\",\"Resource\":\"arn:aws:sns:us-east-1:0000000000:topic\",\"Condition\":{\"StringEquals\":{\"aws:SourceOwner\":\"0000000000\"},\"ArnLike\":{\"aws:SourceArn\":[\"arn:aws:s3:*:*:old_bucket\"]}}}]}"),
				},
			},
			mockGetTopicAttributesError: nil,
			mockSetTopicAttributesError: errors.New("mock set topic attributes error"),
			add:                         true,
			attributeValue:              nil,
			error:                       "mock set topic attributes error",
		},
		{
			description: "successful add topic policy bucket arn invocation",
			mockGetTopicAttributesOutput: &sns.GetTopicAttributesOutput{
				Attributes: map[string]*string{
					"Policy": aws.String("{\"Version\":\"2008-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"s3.amazonaws.com\"},\"Action\":\"sns:Publish\",\"Resource\":\"arn:aws:sns:us-east-1:0000000000:topic\",\"Condition\":{\"StringEquals\":{\"aws:SourceOwner\":\"0000000000\"},\"ArnLike\":{\"aws:SourceArn\":[\"arn:aws:s3:*:*:old_bucket\"]}}}]}"),
				},
			},
			mockGetTopicAttributesError: nil,
			mockSetTopicAttributesError: nil,
			add:                         true,
			attributeValue:              aws.String(`{"Version":"2008-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"s3.amazonaws.com"},"Action":"sns:Publish","Resource":"arn:aws:sns:us-east-1:0000000000:topic","Condition":{"StringEquals":{"aws:SourceOwner":"0000000000"},"ArnLike":{"aws:SourceArn":["arn:aws:s3:*:*:old_bucket","arn:aws:s3:*:*:new_bucket"]}}}]}`),
			error:                       "",
		},
		{
			description: "successful remove topic policy bucket arn invocation",
			mockGetTopicAttributesOutput: &sns.GetTopicAttributesOutput{
				Attributes: map[string]*string{
					"Policy": aws.String("{\"Version\":\"2008-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"s3.amazonaws.com\"},\"Action\":\"sns:Publish\",\"Resource\":\"arn:aws:sns:us-east-1:0000000000:topic\",\"Condition\":{\"StringEquals\":{\"aws:SourceOwner\":\"0000000000\"},\"ArnLike\":{\"aws:SourceArn\":[\"arn:aws:s3:*:*:old_bucket\",\"arn:aws:s3:*:*:new_bucket\"]}}}]}"),
				},
			},
			mockGetTopicAttributesError: nil,
			mockSetTopicAttributesError: nil,
			add:                         false,
			attributeValue:              aws.String(`{"Version":"2008-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"s3.amazonaws.com"},"Action":"sns:Publish","Resource":"arn:aws:sns:us-east-1:0000000000:topic","Condition":{"StringEquals":{"aws:SourceOwner":"0000000000"},"ArnLike":{"aws:SourceArn":["arn:aws:s3:*:*:old_bucket"]}}}]}`),
			error:                       "",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &help{
				topicARN: "new_topic_arn",
				snsClient: &mockSNSClient{
					mockGetTopicAttributesOutput: test.mockGetTopicAttributesOutput,
					mockGetTopicAttributesError:  test.mockGetTopicAttributesError,
					mockSetTopicAttributesError:  test.mockSetTopicAttributesError,
				},
			}

			err := h.addOrRemoveTopicPolicyBucketARN(context.Background(), "new_bucket", test.add)

			if err != nil {
				if err.Error() != test.error {
					t.Errorf("incorrect error, received: %s, expected: %s", err.Error(), test.error)
				}
			} else {
				receivedAttributeValue := h.snsClient.(*mockSNSClient).attributeValue
				if *receivedAttributeValue != *test.attributeValue {
					t.Errorf("incorrect attribute value, received: %s, expected: %s", *receivedAttributeValue, *test.attributeValue)
				}
			}
		})
	}
}
