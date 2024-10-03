package rest_api

import (
	"context"
	"encore.app/bill/api_flow"
	"encore.app/bill/models"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"time"
)

//encore:api public method=POST path=/filter_bills
func (s *Service) FilterBills(ctx context.Context, status *models.FilterBillParam) (*models.BillResponce, error) {
	rows, err := models.DB.Query(` SELECT bills.*,lineitems.amount,lineitems.itemdate  FROM bills left join lineitems on bills.id = lineitems.bill_id where bills.status=($1)`, status.Status)
	if err != nil {
		return nil, err
	}

	dict := map[string]models.Bill{}

	for rows.Next() {
		var bill models.Bill
		var litem models.LineItem
		err := rows.Scan(&bill.ID, &bill.Status, &bill.CreatedAt, &bill.ClosedAt, &bill.Currency, &bill.Total, &litem.Amount, &litem.CreatedAt)
		if err == nil {
			value, ok := dict[bill.ID]
			if ok {
				if litem.CreatedAt.Valid {
					value.LineItems = append(value.LineItems, models.LineItem{
						BillID:    bill.ID,
						Amount:    litem.Amount,
						CreatedAt: litem.CreatedAt,
					})
					dict[bill.ID] = value
				}

			} else {
				billst := models.Bill{
					ID:        bill.ID,
					Status:    bill.Status,
					CreatedAt: bill.CreatedAt,
					ClosedAt:  bill.ClosedAt,
					Currency:  bill.Currency,
					Total:     bill.Total,
					LineItems: []models.LineItem{},
				}
				if litem.Amount.Valid && litem.CreatedAt.Valid {
					billst.LineItems = []models.LineItem{
						{
							BillID:    bill.ID,
							Amount:    litem.Amount,
							CreatedAt: litem.CreatedAt,
						},
					}
				}

				dict[bill.ID] = billst
			}

		}

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &models.BillResponce{
		Bills: dict,
	}, err

}

//encore:api public method=POST path=/bill
func (s *Service) AddBill(ctx context.Context, p *models.AddBillParam) (*models.Response, error) {
	BillId := uuid.New().String()
	options := client.StartWorkflowOptions{
		ID:        "BillWorkflow-" + BillId,
		TaskQueue: BillTaskQueue,

		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // Disabling retries
		}}

	_, err := s.client.ExecuteWorkflow(ctx, options, api_flow.BillWorkflow, p, BillId)
	if err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Second)
	response, err := s.client.QueryWorkflow(ctx, "BillWorkflow-"+BillId, "", "getRes")
	if err != nil {
		return nil, errs.Wrap(err, "Failed to query workflow")
	}
	var res models.Response
	err = response.Get(&res)

	return &res, err
}

//encore:api public method=POST path=/line_item
func (s *Service) AddLineItem(ctx context.Context, p *models.LineParam) (*models.Response, error) {

	err := s.client.SignalWorkflow(ctx, "BillWorkflow-"+p.BillID, "", "addLineItem", p)
	if err != nil {
		return &models.Response{
			Message: "BillId is invalid or bill is already closed",
			Code:    1,
		}, nil
	}
	time.Sleep(1 * time.Second)

	response, err := s.client.QueryWorkflow(ctx, "BillWorkflow-"+p.BillID, "", "getRes")
	if err != nil {
		return nil, errs.Wrap(err, "Failed to query workflow")
	}
	var res models.Response
	err = response.Get(&res)

	return &res, nil
}

//encore:api public method=PATCH path=/bill/:BillID
func (s *Service) CloseBill(ctx context.Context, BillID string) (*models.Response, error) {

	err := s.client.SignalWorkflow(ctx, "BillWorkflow-"+BillID, "", "closeBill", BillID)
	if err != nil {
		return &models.Response{
			Message: "BillId is invalid or bill is already closed",
			Code:    1,
		}, nil
	}
	time.Sleep(1 * time.Second)

	response, err := s.client.QueryWorkflow(ctx, "BillWorkflow-"+BillID, "", "getRes")
	if err != nil {
		return nil, errs.Wrap(err, "Failed to query workflow")
	}
	var res models.Response
	err = response.Get(&res)

	return &res, nil
}
