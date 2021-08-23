package fs

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
)

type mockHelper struct {
	mockAddOrRemoveNotificationError error
}

func (m *mockHelper) addOrRemoveNotification(ctx context.Context, path string, add bool) error {
	return m.mockAddOrRemoveNotificationError
}

func TestNew(t *testing.T) {
	client := New(session.New(), "topic arn")
	if client == nil {
		t.Error("error creating filesystem client")
	}
}

func TestCreateFileWatcher(t *testing.T) {
	tests := []struct {
		description                      string
		mockAddOrRemoveNotificationError error
		error                            error
	}{
		{
			description:                      "error adding notification",
			mockAddOrRemoveNotificationError: errors.New("mock add notification error"),
			error:                            &ErrorAddNotification{},
		},
		{
			description:                      "successful create file watcher invocation",
			mockAddOrRemoveNotificationError: nil,
			error:                            nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockAddOrRemoveNotificationError: test.mockAddOrRemoveNotificationError,
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
		description                      string
		mockAddOrRemoveNotificationError error
		error                            error
	}{
		{
			description:                      "error removing notification",
			mockAddOrRemoveNotificationError: errors.New("mock remove notification error"),
			error:                            &ErrorAddNotification{},
		},
		{
			description:                      "successful delete file watcher invocation",
			mockAddOrRemoveNotificationError: nil,
			error:                            nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelper{
					mockAddOrRemoveNotificationError: test.mockAddOrRemoveNotificationError,
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
