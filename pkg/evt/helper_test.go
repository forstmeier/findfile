package evt

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

type mockCloudTrailClient struct {
	mockGetEventSelectorsOutput *cloudtrail.GetEventSelectorsOutput
	mockGetEventSelectorsError  error
	mockPutEventSelectorsOutput *cloudtrail.PutEventSelectorsOutput
	mockPutEventSelectorsError  error
}

func (mc *mockCloudTrailClient) GetEventSelectors(input *cloudtrail.GetEventSelectorsInput) (*cloudtrail.GetEventSelectorsOutput, error) {
	return mc.mockGetEventSelectorsOutput, mc.mockGetEventSelectorsError
}

func (mc *mockCloudTrailClient) PutEventSelectors(input *cloudtrail.PutEventSelectorsInput) (*cloudtrail.PutEventSelectorsOutput, error) {
	return mc.mockPutEventSelectorsOutput, mc.mockPutEventSelectorsError
}

func Test_getEventValues(t *testing.T) {
	mockError := errors.New("mock get event selectors")
	arn := "arn:aws:s3:::bucket"

	tests := []struct {
		description                 string
		mockGetEventSelectorsOutput *cloudtrail.GetEventSelectorsOutput
		mockGetEventSelectorsError  error
		values                      []*string
		error                       error
	}{
		{
			description:                 "error getting event selectors",
			mockGetEventSelectorsOutput: nil,
			mockGetEventSelectorsError:  mockError,
			values:                      nil,
			error:                       mockError,
		},
		{
			description:                 "no event selectors received",
			mockGetEventSelectorsOutput: &cloudtrail.GetEventSelectorsOutput{},
			mockGetEventSelectorsError:  nil,
			values:                      []*string{},
			error:                       nil,
		},
		{
			description: "no data resources received",
			mockGetEventSelectorsOutput: &cloudtrail.GetEventSelectorsOutput{
				EventSelectors: []*cloudtrail.EventSelector{},
			},
			mockGetEventSelectorsError: nil,
			values:                     []*string{},
			error:                      nil,
		},
		{
			description: "no values received",
			mockGetEventSelectorsOutput: &cloudtrail.GetEventSelectorsOutput{
				EventSelectors: []*cloudtrail.EventSelector{
					{
						DataResources: []*cloudtrail.DataResource{},
					},
				},
			},
			mockGetEventSelectorsError: nil,
			values:                     []*string{},
			error:                      nil,
		},
		{
			description: "successful invocation",
			mockGetEventSelectorsOutput: &cloudtrail.GetEventSelectorsOutput{
				EventSelectors: []*cloudtrail.EventSelector{
					{
						DataResources: []*cloudtrail.DataResource{
							{
								Values: []*string{&arn},
							},
						},
					},
				},
			},
			mockGetEventSelectorsError: nil,
			values: []*string{
				&arn,
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := &mockCloudTrailClient{
				mockGetEventSelectorsOutput: test.mockGetEventSelectorsOutput,
				mockGetEventSelectorsError:  test.mockGetEventSelectorsError,
			}

			h := &help{
				cloudtrailClient: c,
			}

			values, err := h.getEventValues("trailName")

			if err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}

			if !reflect.DeepEqual(values, test.values) {
				t.Errorf("incorrect values, received: %v, expected: %v", values, test.values)
			}
		})
	}
}

func Test_putEventValues(t *testing.T) {
	mockError := errors.New("mock put event selectors")

	tests := []struct {
		description                 string
		mockPutEventSelectorsOutput *cloudtrail.PutEventSelectorsOutput
		mockPutEventSelectorsError  error
		error                       error
	}{
		{
			description:                 "error putting event selectors",
			mockPutEventSelectorsOutput: nil,
			mockPutEventSelectorsError:  mockError,
			error:                       mockError,
		},
		{
			description:                 "successful invocation",
			mockPutEventSelectorsOutput: nil,
			mockPutEventSelectorsError:  nil,
			error:                       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			c := &mockCloudTrailClient{
				mockPutEventSelectorsOutput: test.mockPutEventSelectorsOutput,
				mockPutEventSelectorsError:  test.mockPutEventSelectorsError,
			}

			h := &help{
				cloudtrailClient: c,
			}

			err := h.putEventValues("trailName", []*string{})

			if err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}
		})
	}
}
