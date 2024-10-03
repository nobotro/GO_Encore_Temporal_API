package test

import (
	"context"
	"encore.app/bill/api_flow"
	"encore.app/bill/models"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"testing"
)

var DB = sqldb.NewDatabase("fee_api_testing_database", sqldb.DatabaseConfig{
	Migrations: "./migrations"}).Stdlib()

// Test createBillActivity
func TestActivityAddBill(t *testing.T) {
	par := models.AddBillParam{
		Currency: "gel",
	}
	// Call the activity
	billID := "test-bill-id"
	res, _ := api_flow.ActivityAddBill(context.Background(), &par, billID)

	// Assert expectations

	assert.Equal(t, res.Message, "Bill Added Successfully, ID:test-bill-id")
	assert.Equal(t, res.Code, int64(0))

}

func TestActivityAddLineItem(t *testing.T) {

	// Call the activity
	billID := "test-bill-id"
	par := models.LineParam{
		BillID: billID,
		Amount: 10.9,
	}
	par2 := models.LineParam{
		BillID: "99999999999999999999999999999",
		Amount: 10.9,
	}
	res, _ := api_flow.ActivityAddLineItem(context.Background(), &par)
	assert.Equal(t, res.Message, "Line Item Added Successfully")
	assert.Equal(t, res.Code, int64(0))

	res, _ = api_flow.ActivityAddLineItem(context.Background(), &par2)
	assert.Equal(t, res.Message, "Line Item Add failed")
	assert.Equal(t, res.Code, int64(1))

}
func TestActivityActivityCloseBill(t *testing.T) {

	// Call the activity
	billID := "test-bill-id"
	res, _ := api_flow.ActivityCloseBill(context.Background(), billID)
	expected_responce := "Bill Closed Successfully\n Total Amount Charged: 10.900000 \n Line Items: \n Charged amount : 10.900000\n"
	assert.Equal(t, res.Message, expected_responce)
	assert.Equal(t, res.Code, int64(0))
	DB.Exec("drop database fee_api_testing_database")
}
