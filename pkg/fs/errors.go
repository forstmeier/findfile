package fs

import "fmt"

const packageName = "fs"

// ErrorAddNotification wraps errors returned by
// helper.addOrRemoveNotification.
type ErrorAddNotification struct {
	err error
}

func (e *ErrorAddNotification) Error() string {
	return fmt.Sprintf("[%s] [create file watcher] [add notification]: %s", packageName, e.err.Error())
}

// ErrorAddTopicPolicyBucketARN wraps errors returned by
// helper.addOrRemoveTopicPolicyBucketARN.
type ErrorAddTopicPolicyBucketARN struct {
	err error
}

func (e *ErrorAddTopicPolicyBucketARN) Error() string {
	return fmt.Sprintf("[%s] [create file watcher] [add topic policy bucket arn]: %s", packageName, e.err.Error())
}

// ErrorRemoveNotification wraps errors returned by
// helper.addOrRemoveNotification.
type ErrorRemoveNotification struct {
	err error
}

func (e *ErrorRemoveNotification) Error() string {
	return fmt.Sprintf("[%s] [delete file watcher] [remove notification]: %s", packageName, e.err.Error())
}

// ErrorRemoveTopicPolicyBucketARN wraps errors returned by
// helper.addOrRemoveTopicPolicyBucketARN.
type ErrorRemoveTopicPolicyBucketARN struct {
	err error
}

func (e *ErrorRemoveTopicPolicyBucketARN) Error() string {
	return fmt.Sprintf("[%s] [delete file watcher] [remove topic policy bucket arn]: %s", packageName, e.err.Error())
}
