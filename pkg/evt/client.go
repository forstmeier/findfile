package evt

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
)

const arnPrefix = "arn:aws:s3:::"

var _ Eventer = &Client{}

// Client implements the evt.Eventer methods using AWS CloudTrail.
type Client struct {
	trailName string
	helper    helper
}

// New generates an evt.Client pointer instance with AWS CloudTrail.
func New(newSession *session.Session, trailName string) *Client {
	return &Client{
		trailName: trailName,
		helper: &help{
			cloudtrailClient: cloudtrail.New(newSession),
		},
	}
}

// AddBucketListeners implements the evt.Eventer.AddBucketListeners method
// using AWS CloudTrail.
func (c *Client) AddBucketListeners(ctx context.Context, buckets []string) error {
	values, err := c.helper.getEventValues(c.trailName)
	if err != nil {
		return &GetEventValuesError{
			err: err,
		}
	}

	valuesMap := map[string]struct{}{}
	for _, value := range values {
		valuesMap[*value] = struct{}{}
	}

	for _, bucket := range buckets {
		newARN := arnPrefix + bucket + "/"
		if _, ok := valuesMap[newARN]; !ok {
			values = append(values, &newARN)
		}
	}

	if err := c.helper.putEventValues(c.trailName, values); err != nil {
		return &PutEventValuesError{
			err: err,
		}
	}

	return nil
}

// RemoveBucketListeners implements the evt.Eventer.RemoveBucketListeners
// method using AWS CloudTrail.
func (c *Client) RemoveBucketListeners(ctx context.Context, buckets []string) error {
	values, err := c.helper.getEventValues(c.trailName)
	if err != nil {
		return &GetEventValuesError{
			err: err,
		}
	}

	bucketsMap := map[string]struct{}{}
	for _, bucket := range buckets {
		bucketsMap[arnPrefix+bucket+"/"] = struct{}{}
	}

	for i, value := range values {
		if _, ok := bucketsMap[*value]; ok {
			values = append(values[:i], values[i+1:]...)
		}
	}

	if err := c.helper.putEventValues(c.trailName, values); err != nil {
		return &PutEventValuesError{
			err: err,
		}
	}

	return nil
}
