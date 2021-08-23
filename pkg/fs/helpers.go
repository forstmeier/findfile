package fs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type helper interface {
	addOrRemoveNotification(ctx context.Context, path string, add bool) error
}

type help struct {
	topicARN string
	s3Client s3Client
}

type s3Client interface {
	GetBucketNotificationConfiguration(input *s3.GetBucketNotificationConfigurationRequest) (*s3.NotificationConfiguration, error)
	PutBucketNotificationConfiguration(input *s3.PutBucketNotificationConfigurationInput) (*s3.PutBucketNotificationConfigurationOutput, error)
}

func (h *help) addOrRemoveNotification(ctx context.Context, path string, add bool) error {
	config, err := h.s3Client.GetBucketNotificationConfiguration(&s3.GetBucketNotificationConfigurationRequest{
		Bucket: aws.String(path),
	})
	if err != nil {
		return err
	}

	if add {
		config.TopicConfigurations = append(config.TopicConfigurations, &s3.TopicConfiguration{
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
		Bucket:                    aws.String(path),
		NotificationConfiguration: config,
	})
	if err != nil {
		return err
	}

	return nil
}
