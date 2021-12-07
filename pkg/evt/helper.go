package evt

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

type helper interface {
	getEventValues(trailName string) ([]*string, error)
	putEventValues(trailName string, values []*string) error
}

type help struct {
	cloudtrailClient cloudtrailClient
}

type cloudtrailClient interface {
	GetEventSelectors(input *cloudtrail.GetEventSelectorsInput) (*cloudtrail.GetEventSelectorsOutput, error)
	PutEventSelectors(input *cloudtrail.PutEventSelectorsInput) (*cloudtrail.PutEventSelectorsOutput, error)
}

func (h *help) getEventValues(trailName string) ([]*string, error) {
	output, err := h.cloudtrailClient.GetEventSelectors(&cloudtrail.GetEventSelectorsInput{
		TrailName: &trailName,
	})
	if err != nil {
		return nil, err
	}

	if len(output.EventSelectors) == 0 {
		return []*string{}, nil
	}
	if len(output.EventSelectors[0].DataResources) == 0 {
		return []*string{}, nil
	}
	if len(output.EventSelectors[0].DataResources[0].Values) == 0 {
		return []*string{}, nil
	}

	return output.EventSelectors[0].DataResources[0].Values, nil
}

func (h *help) putEventValues(trailName string, values []*string) error {
	_, err := h.cloudtrailClient.PutEventSelectors(&cloudtrail.PutEventSelectorsInput{
		TrailName: &trailName,
		EventSelectors: []*cloudtrail.EventSelector{
			{
				DataResources: []*cloudtrail.DataResource{
					{
						Type:   aws.String("AWS::S3::Object"),
						Values: values,
					},
				},
				ReadWriteType: aws.String(cloudtrail.ReadWriteTypeWriteOnly),
			},
		},
	})

	return err
}
