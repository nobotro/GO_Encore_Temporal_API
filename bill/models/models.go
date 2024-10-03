package models

import (
	"database/sql"
	"encore.dev/storage/sqldb"
	"fmt"
)

type BillStatus string
type Currency string

const (
	OPEN   BillStatus = "open"
	CLOSED BillStatus = "closed"
	USD    Currency   = "usd"
	GEL    Currency   = "gel"
)

type Bill struct {
	ID        string
	Status    BillStatus
	CreatedAt sql.NullTime
	ClosedAt  sql.NullTime
	Currency  Currency
	Total     float64
	LineItems []LineItem
}

type LineItem struct {
	BillID    string
	Amount    sql.NullFloat64
	CreatedAt sql.NullTime
}

type AddBillParam struct {
	Currency Currency
}

type FilterBillParam struct {
	Status BillStatus
}

type LineParam struct {
	BillID string
	Amount float64
}

type Response struct {
	Message string
	Code    int64
}

type BillResponce struct {
	Bills map[string]Bill
}

func (l LineItem) String() string {
	return fmt.Sprintf(" Charged amount : %f", l.Amount.Float64)
}

var DB = sqldb.NewDatabase("fee_api", sqldb.DatabaseConfig{
	Migrations: "./migrations"}).Stdlib()
