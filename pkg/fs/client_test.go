package fs

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
)

type mockHelper struct {
	mockAddOrRemoveNotificationError         error
	mockAddOrRemoveTopicPolicyBucketARNError error
}

func (m *mockHelper) addOrRemoveNotification(ctx context.Context, bucketName string, add bool) error {
	return m.mockAddOrRemoveNotificationError
}

func (m *mockHelper) addOrRemoveTopicPolicyBucketARN(ctx context.Context, bucketName string, add bool) error {
	return m.mockAddOrRemoveTopicPolicyBucketARNError
}

func TestNew(t *testing.T) {
	client := New(session.New(), "topic_arn", "configuration_id")
	if client == nil {
		t.Error("error creating filesystem client")
	}
}

func TestCreateFileWatcher(t *testing.T) {
	tests := []struct {
		description                              string
		mockAddOrRemoveNotificationError         error
		mockAddOrRemoveTopicPolicyBucketARNError error
		error                                    error
	}{
		{
			description:                              "error adding notification",
			mockAddOrRemoveNotificationError:         errors.New("mock add notification error"),
			mockAddOrRemoveTopicPolicyBucketARNError: nil,
			error:                                    &ErrorAddNotification{},
		},
		{
			description:                              "error adding topic policy bucket arn",
			mockAddOrRemoveNotificationError:         nil,
			mockAddOrRemoveTopicPolicyBucketARNError: errors.New("mock add topic policy bucket arn error"),
			error:                                    &ErrorAddTopicPolicyBucketARN{},
		},
		{
			description:                              "successful create file watcher invocation",
			mockAddOrRemoveNotificationError:         nil,
			mockAddOrRemoveTopicPolicyBucketARNError: nil,
			error:                                    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockAddOrRemoveNotificationError:         test.mockAddOrRemoveNotificationError,
					mockAddOrRemoveTopicPolicyBucketARNError: test.mockAddOrRemoveTopicPolicyBucketARNError,
				},
			}

			err := client.CreateFileWatcher(context.Background(), "path")

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorAddNotification:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}

func TestDeleteFileWatcher(t *testing.T) {
	tests := []struct {
		description                              string
		mockAddOrRemoveNotificationError         error
		mockAddOrRemoveTopicPolicyBucketARNError error
		error                                    error
	}{
		{
			description:                              "error removing notification",
			mockAddOrRemoveNotificationError:         errors.New("mock remove notification error"),
			mockAddOrRemoveTopicPolicyBucketARNError: nil,
			error:                                    &ErrorAddNotification{},
		},
		{
			description:                              "error removing topic policy bucket arn",
			mockAddOrRemoveNotificationError:         nil,
			mockAddOrRemoveTopicPolicyBucketARNError: errors.New("mock remove topic policy bucket arn error"),
			error:                                    &ErrorAddNotification{},
		},
		{
			description:                              "successful delete file watcher invocation",
			mockAddOrRemoveNotificationError:         nil,
			mockAddOrRemoveTopicPolicyBucketARNError: nil,
			error:                                    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockAddOrRemoveNotificationError:         test.mockAddOrRemoveNotificationError,
					mockAddOrRemoveTopicPolicyBucketARNError: test.mockAddOrRemoveTopicPolicyBucketARNError,
				},
			}

			err := client.DeleteFileWatcher(context.Background(), "path")

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorRemoveNotification:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				}
			} else {
				if err != test.error {
					t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
				}
			}
		})
	}
}
