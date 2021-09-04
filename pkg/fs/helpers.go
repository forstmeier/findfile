package fs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

type helper interface {
	addOrRemoveNotification(ctx context.Context, bucketName string, add bool) error
	addOrRemoveTopicPolicyBucketARN(ctx context.Context, bucketName string, add bool) error
}

type help struct {
	topicARN        string
	configurationID string
	s3Client        s3Client
	snsClient       snsClient
}

type s3Client interface {
	GetBucketNotificationConfiguration(input *s3.GetBucketNotificationConfigurationRequest) (*s3.NotificationConfiguration, error)
	PutBucketNotificationConfiguration(input *s3.PutBucketNotificationConfigurationInput) (*s3.PutBucketNotificationConfigurationOutput, error)
}

type snsClient interface {
	GetTopicAttributes(input *sns.GetTopicAttributesInput) (*sns.GetTopicAttributesOutput, error)
	SetTopicAttributes(input *sns.SetTopicAttributesInput) (*sns.SetTopicAttributesOutput, error)
}

// addOrRemoveNotification updates the target user S3 bucket with the required
// notification configuration
func (h *help) addOrRemoveNotification(ctx context.Context, bucketName string, add bool) error {
	config, err := h.s3Client.GetBucketNotificationConfiguration(&s3.GetBucketNotificationConfigurationRequest{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}

	if add {
		config.TopicConfigurations = append(config.TopicConfigurations, &s3.TopicConfiguration{
			Id: aws.String(h.configurationID),
			Events: []*string{
				aws.String("s3:ObjectCreated:*"),
				aws.String("s3:ObjectRemoved:*"),
			},
			TopicArn: aws.String(h.topicARN),
		})
	} else {
		index := 0
		for i, topicConfig := range config.TopicConfigurations {
			if *topicConfig.TopicArn == h.topicARN {
				index = i
				break
			}
		}
		config.TopicConfigurations = append(config.TopicConfigurations[:index], config.TopicConfigurations[index+1:]...)
	}

	_, err = h.s3Client.PutBucketNotificationConfiguration(&s3.PutBucketNotificationConfigurationInput{
		Bucket:                    aws.String(bucketName),
		NotificationConfiguration: config,
	})
	if err != nil {
		return err
	}

	return nil
}

type policy struct {
	Version   string      `json:"Version"`
	Statement []statement `json:"Statement"`
}

type statement struct {
	Effect    string            `json:"Effect"`
	Principal map[string]string `json:"Principal"`
	Action    string            `json:"Action"`
	Resource  string            `json:"Resource"`
	Condition condition         `json:"Condition"`
}

type condition struct {
	StringEquals map[string]string   `json:"StringEquals"`
	ArnLike      map[string][]string `json:"ArnLike"`
}

// addOrRemoveTopicPolicyBucketARN updates the SNS topic policy with the ARN of
// the provided bucketName.
func (h *help) addOrRemoveTopicPolicyBucketARN(ctx context.Context, bucketName string, add bool) error {
	topicAttributes, err := h.snsClient.GetTopicAttributes(&sns.GetTopicAttributesInput{
		TopicArn: aws.String(h.topicARN),
	})
	if err != nil {
		return err
	}

	// a policy is guaranteed to exist on the SNS topic due to the
	// CloudFormation template
	policyString := topicAttributes.Attributes["Policy"]

	policyJSON := policy{}
	if err := json.Unmarshal([]byte(*policyString), &policyJSON); err != nil {
		return err
	}

	arn := fmt.Sprintf("arn:aws:s3:*:*:%s", bucketName)
	arns := policyJSON.Statement[0].Condition.ArnLike["aws:SourceArn"]
	if add {
		arns = append(arns, arn)
		policyJSON.Statement[0].Condition.ArnLike["aws:SourceArn"] = arns
	} else {
		for i, policyARN := range arns {
			if policyARN == arn {
				arns = append(arns[:i], arns[i+1:]...)
				policyJSON.Statement[0].Condition.ArnLike["aws:SourceArn"] = arns
				break
			}
		}
	}

	policyBytes, err := json.Marshal(policyJSON)
	if err != nil {
		return err
	}

	_, err = h.snsClient.SetTopicAttributes(&sns.SetTopicAttributesInput{
		AttributeName:  aws.String("Policy"),
		AttributeValue: aws.String(string(policyBytes)),
		TopicArn:       aws.String(h.topicARN),
	})
	if err != nil {
		return err
	}

	return nil
}
