package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/cfn"

	"github.com/forstmeier/findfile/pkg/db"
	"github.com/forstmeier/findfile/util"
)

func handler(dbClient db.Databaser, sendResponse func(response *cfn.Response) error) func(ctx context.Context, event cfn.Event) error {
	return func(ctx context.Context, event cfn.Event) error {
		util.Log("EVENT_BODY", event)

		response := cfn.NewResponse(&event)
		response.Status = cfn.StatusSuccess
		response.PhysicalResourceID = "setupCustomResource"

		if event.RequestType != cfn.RequestCreate {
			message := fmt.Sprintf(`received non-create event type [%s]`, event.RequestType)
			util.Log("NON_CREATE_EVENT_ERROR", message)
			response.Reason = message
		} else {
			if err := dbClient.SetupDatabase(ctx); err != nil {
				util.Log("SETUP_DATABASE_ERROR", err.Error())
				response.Reason = fmt.Sprintf("setup database error [%s]", err.Error())
			} else {
				response.Reason = "successful invocation"
			}
		}

		if err := sendResponse(response); err != nil {
			util.Log("SEND_ERROR", err.Error())
			response.Reason = fmt.Sprintf("send response error [%s]", err.Error())
		}

		return nil
	}
}

func sendResponse(response *cfn.Response) error {
	return response.Send()
}
