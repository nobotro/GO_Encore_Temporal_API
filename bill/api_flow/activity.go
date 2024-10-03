package api_flow

import (
	"context"

	"encore.app/bill/models"
	"fmt"
	"time"
)

func ActivityAddBill(ctx context.Context, p *models.AddBillParam, BillId string) (*models.Response, error) {

	if p.Currency != models.USD && p.Currency != models.GEL {

		return &models.Response{Message: "Invalid Currency ,Only gel and usd supported", Code: 1}, nil
	}
	_, err := models.DB.Exec(`INSERT INTO bills (id,currency) VALUES ($1 , $2)`, BillId, p.Currency)
	if err != nil {
		return &models.Response{Message: "Bill Add failed", Code: 1}, nil
	}
	return &models.Response{Message: fmt.Sprintf("Bill Added Successfully, ID:%s", BillId), Code: 0}, nil

}

func ActivityAddLineItem(ctx context.Context, p *models.LineParam) (*models.Response, error) {

	var bill models.Bill

	row := models.DB.QueryRow(`SELECT status FROM bills where id=$1`, p.BillID)
	err := row.Scan(&bill.Status)

	if err == nil && bill.Status == models.OPEN {
		_, err := models.DB.Exec(`INSERT INTO lineitems (bill_id,amount) VALUES ($1 , $2)`, p.BillID, p.Amount)
		if err != nil {
			return &models.Response{Message: "Line Item Add failed", Code: 1}, nil
		}

		_, err = models.DB.Exec(`update bills set total_amount =(select sum(amount) from lineitems where bill_id =$1) where id =$2 `, p.BillID, p.BillID)
		if err != nil {
			return &models.Response{Message: "Line Item Add failed", Code: 1}, nil
		}

	} else {
		return &models.Response{Message: "Line Item Add failed", Code: 1}, nil
	}

	return &models.Response{Message: "Line Item Added Successfully", Code: 0}, nil
}

func ActivityCloseBill(ctx context.Context, id string) (*models.Response, error) {

	var bill models.Bill
	row := models.DB.QueryRow(`SELECT * FROM bills where id=$1`, id)
	err := row.Scan(&bill.ID, &bill.Status, &bill.CreatedAt, &bill.ClosedAt, &bill.Currency, &bill.Total)
	if err != nil {
		return nil, err
	}
	message := ""
	total_amount := 0.0
	if bill.Status == models.OPEN {

		rows, err := models.DB.Query(`SELECT amount, itemdate FROM lineitems where bill_id=($1)`, bill.ID)

		if err != nil {
			return &models.Response{Message: "Bill closing failed", Code: 1}, nil
		}

		for rows.Next() {
			var li models.LineItem
			if err := rows.Scan(&li.Amount, &li.CreatedAt); err != nil {

				return &models.Response{Message: "Bill closing failed", Code: 1}, nil

			}
			message = message + li.String() + "\n"
			total_amount = total_amount + li.Amount.Float64
		}

		_, err = models.DB.Exec(`UPDATE bills set status = 'closed', total_amount = ($2) , closed_at = ($3) where id = ($1) and status = 'open'`, id, total_amount, time.Now())

		if err != nil {
			return &models.Response{Message: "Bill closing failed", Code: 1}, nil
		}
	} else {

		return &models.Response{Message: "Bill already closed", Code: 1}, nil
	}

	return &models.Response{Message: "Bill Closed Successfully\n" +
		fmt.Sprintf(" Total Amount Charged: %f \n", total_amount) +
		" Line Items: \n" + message, Code: 0}, nil
}
