package evt

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type mockHelper struct {
	mockGetEventValuesOutput   []*string
	mockGetEventValuesError    error
	mockPutEventValuesReceived []*string
	mockPutEventValuesError    error
}

func (mh *mockHelper) getEventValues(trailName string) ([]*string, error) {
	return mh.mockGetEventValuesOutput, mh.mockGetEventValuesError
}

func (mh *mockHelper) putEventValues(trailName string, values []*string) error {
	mh.mockPutEventValuesReceived = values
	return mh.mockPutEventValuesError
}

func TestAddBucketListeners(t *testing.T) {
	oldValue := arnPrefix + "old_bucket"
	newBucket := "new_bucket"
	newValue := arnPrefix + newBucket

	tests := []struct {
		description                string
		mockGetEventValuesOutput   []*string
		mockGetEventValuesError    error
		mockPutEventValuesReceived []*string
		mockPutEventValuesError    error
		error                      error
	}{
		{
			description:                "get event values error",
			mockGetEventValuesOutput:   nil,
			mockGetEventValuesError:    errors.New("mock get event values error"),
			mockPutEventValuesReceived: nil,
			mockPutEventValuesError:    nil,
			error:                      &GetEventValuesError{},
		},
		{
			description:                "put event values error",
			mockGetEventValuesOutput:   nil,
			mockGetEventValuesError:    nil,
			mockPutEventValuesReceived: nil,
			mockPutEventValuesError:    errors.New("mock put event values error"),
			error:                      &PutEventValuesError{},
		},
		{
			description:                "successful invocation",
			mockGetEventValuesOutput:   []*string{&oldValue},
			mockGetEventValuesError:    nil,
			mockPutEventValuesReceived: []*string{&oldValue, &newValue},
			mockPutEventValuesError:    nil,
			error:                      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockGetEventValuesOutput: test.mockGetEventValuesOutput,
				mockGetEventValuesError:  test.mockGetEventValuesError,
				mockPutEventValuesError:  test.mockPutEventValuesError,
			}

			c := &Client{
				trailName: "trailName",
				helper:    h,
			}

			err := c.AddBucketListeners(context.Background(), []string{newBucket})

			if err != nil {
				switch e := test.error.(type) {
				case *GetEventValuesError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				case *PutEventValuesError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if !reflect.DeepEqual(c.helper.(*mockHelper).mockPutEventValuesReceived, test.mockPutEventValuesReceived) {
					t.Errorf("incorrect values, received: %v, expected: %v", h.mockPutEventValuesReceived, test.mockPutEventValuesReceived)
				}
			}
		})
	}
}

func TestRemoveBucketListeners(t *testing.T) {
	oldValue := arnPrefix + "old_bucket"
	removeBucket := "remove_bucket"
	removeValue := arnPrefix + removeBucket

	tests := []struct {
		description                string
		mockGetEventValuesOutput   []*string
		mockGetEventValuesError    error
		mockPutEventValuesReceived []*string
		mockPutEventValuesError    error
		error                      error
	}{
		{
			description:                "get event values error",
			mockGetEventValuesOutput:   nil,
			mockGetEventValuesError:    errors.New("mock get event values error"),
			mockPutEventValuesReceived: nil,
			mockPutEventValuesError:    nil,
			error:                      &GetEventValuesError{},
		},
		{
			description:                "put event values error",
			mockGetEventValuesOutput:   nil,
			mockGetEventValuesError:    nil,
			mockPutEventValuesReceived: nil,
			mockPutEventValuesError:    errors.New("mock put event values error"),
			error:                      &PutEventValuesError{},
		},
		{
			description:                "successful invocation",
			mockGetEventValuesOutput:   []*string{&oldValue, &removeValue},
			mockGetEventValuesError:    nil,
			mockPutEventValuesReceived: []*string{&oldValue},
			mockPutEventValuesError:    nil,
			error:                      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			h := &mockHelper{
				mockGetEventValuesOutput: test.mockGetEventValuesOutput,
				mockGetEventValuesError:  test.mockGetEventValuesError,
				mockPutEventValuesError:  test.mockPutEventValuesError,
			}

			c := &Client{
				trailName: "trailName",
				helper:    h,
			}

			err := c.RemoveBucketListeners(context.Background(), []string{removeBucket})

			if err != nil {
				switch e := test.error.(type) {
				case *GetEventValuesError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				case *PutEventValuesError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if !reflect.DeepEqual(c.helper.(*mockHelper).mockPutEventValuesReceived, test.mockPutEventValuesReceived) {
					t.Errorf("incorrect values, received: %v, expected: %v", h.mockPutEventValuesReceived, test.mockPutEventValuesReceived)
				}
			}
		})
	}
}
