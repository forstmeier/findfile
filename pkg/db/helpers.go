package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (c *Client) uploadObject(ctx context.Context, body interface{}, key, entity string) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = c.s3Client.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(bodyBytes)),
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) listDocumentKeys(ctx context.Context, bucket, prefix string) ([]string, error) {
	var results []string
	var continuationToken string
	for {
		output, err := c.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: aws.String(continuationToken),
		})
		if err != nil {
			return nil, err
		}

		for _, content := range output.Contents {
			results = append(results, *content.Key)
		}

		if *output.IsTruncated {
			continuationToken = *output.NextContinuationToken
		} else {
			break
		}
	}

	return results, nil
}

func (c *Client) executeQuery(ctx context.Context, query []byte) (*string, *string, error) {
	queryInput := athena.StartQueryExecutionInput{
		QueryString: aws.String(string(query)),
		QueryExecutionContext: &athena.QueryExecutionContext{
			Database: aws.String(c.databaseName),
		},
		ResultConfiguration: &athena.ResultConfiguration{
			OutputLocation: aws.String(fmt.Sprintf("s3://%s/results", c.bucketName)),
		},
	}

	queryOutput, err := c.athenaClient.StartQueryExecution(&queryInput)
	if err != nil {
		return nil, nil, err
	}

	executionInput := &athena.GetQueryExecutionInput{
		QueryExecutionId: queryOutput.QueryExecutionId,
	}
	executionOutput := &athena.GetQueryExecutionOutput{}
	for {
		executionOutput, err = c.athenaClient.GetQueryExecution(executionInput)
		if err != nil {
			return nil, nil, err
		}
		if *executionOutput.QueryExecution.Status.State != "RUNNING" {
			break
		}
		time.Sleep(1 * time.Second)
	}

	return queryOutput.QueryExecutionId, executionOutput.QueryExecution.Status.State, nil
}
