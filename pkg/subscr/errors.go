package subscr

import (
	"fmt"
	"strings"
)

const packageName = "subscr"

// ErrorMissingFields is returned if any SubscriptionInfo fields are
// not populated.
type ErrorMissingFields struct {
	fields []string
}

func (e *ErrorMissingFields) Error() string {
	ids := strings.Join(e.fields, ", ")
	return fmt.Sprintf("[%s] [create subscription] [check info fields]: fields [%s] missing", packageName, ids)
}

// ErrorNewPaymentMethod wraps errors returned by stripe.paymentmethod.New
// in the CreateSubscription method.
type ErrorNewPaymentMethod struct {
	err error
}

func (e *ErrorNewPaymentMethod) Error() string {
	return fmt.Sprintf("[%s] [create subscription] [new payment method]: %s", packageName, e.err.Error())
}

// ErrorNewCustomer wraps errors returned by stripe.customer.New
// in the CreateSubscription method.
type ErrorNewCustomer struct {
	err error
}

func (e *ErrorNewCustomer) Error() string {
	return fmt.Sprintf("[%s] [create subscription] [new customer]: %s", packageName, e.err.Error())
}

// ErrorNewSubscription wraps errors returned by stripe.subscription.New
// in the CreateSubscription method.
type ErrorNewSubscription struct {
	err error
}

func (e *ErrorNewSubscription) Error() string {
	return fmt.Sprintf("[%s] [create subscription] [new subscription]: %s", packageName, e.err.Error())
}

// ErrorDeleteCustomer wraps errors returned by stripe.customer.Del
// in the RemoveSubscription method.
type ErrorDeleteCustomer struct {
	err error
}

func (e *ErrorDeleteCustomer) Error() string {
	return fmt.Sprintf("[%s] [remove subscription] [delete customer]: %s", packageName, e.err.Error())
}

// ErrorCreateUsageRecord wraps errors returned by stripe.usagerecord.New
// in the AddUsage method.
type ErrorCreateUsageRecord struct {
	err error
}

func (e *ErrorCreateUsageRecord) Error() string {
	return fmt.Sprintf("[%s] [add usage] [new usage record]: %s", packageName, e.err.Error())
}
