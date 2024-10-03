package api_flow

import (
	"encore.app/bill/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func BillWorkflow(ctx workflow.Context, p *models.AddBillParam, BillId string) error {

	logger := workflow.GetLogger(ctx)
	// Activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // Disabling retries
		}}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var res models.Response
	workflow.SetQueryHandler(ctx, "getRes", func() (*models.Response, error) {
		return &res, nil
	})

	// Create the bill

	workflow.ExecuteActivity(ctx, ActivityAddBill, p, BillId).Get(ctx, &res)
	billingPeriodTimer := workflow.NewTimer(ctx, 30*24*time.Hour)
	status := true
	// Main event loop
	for {

		selector := workflow.NewSelector(ctx)

		// Handle adding line items
		selector.AddReceive(workflow.GetSignalChannel(ctx, "addLineItem"), func(ch workflow.ReceiveChannel, more bool) {
			res = models.Response{}
			var p models.LineParam
			ch.Receive(ctx, &p)
			workflow.ExecuteActivity(ctx, ActivityAddLineItem, &p).Get(ctx, &res)

		})

		// Handle close bil
		selector.AddReceive(workflow.GetSignalChannel(ctx, "closeBill"), func(ch workflow.ReceiveChannel, more bool) {
			res = models.Response{}
			var id string
			ch.Receive(ctx, &id)
			workflow.ExecuteActivity(ctx, ActivityCloseBill, id).Get(ctx, &res)
			status = false
		})

		// Handle billing period expiration
		selector.AddFuture(billingPeriodTimer, func(f workflow.Future) {
			workflow.ExecuteActivity(ctx, ActivityCloseBill, BillId).Get(ctx, nil)

			status = false
		})

		selector.Select(ctx)
		if status == false {
			workflow.CompleteSession(ctx)
			break
		}

	}
	logger.Info("BillWorkflow completed")
	return nil
}
