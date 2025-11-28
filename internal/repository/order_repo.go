package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderHistoryItem struct {
	OrderID       int
	TotalPrice    float64
	OrderDate     time.Time
	PaymentStatus string
	PaidAt        *time.Time
}

func GetOrderHistory(ctx context.Context, db *pgxpool.Pool, customerID int) ([]OrderHistoryItem, error) {
	query := `
        SELECT 
            o.orderid,
            o.totalprice,
            o.orderdate,
            COALESCE(p.paymentstatus, 'Unpaid') AS paymentstatus,
            p.paidat
        FROM orders o
        LEFT JOIN payments p ON p.orderid = o.orderid
        WHERE o.customerid = $1
          AND o.totalprice > 0        -- completed orders only
          AND o.deleted_at IS NULL
        ORDER BY o.orderdate ASC;
    `

	rows, err := db.Query(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []OrderHistoryItem{}

	for rows.Next() {
		var item OrderHistoryItem
		var paidAt *time.Time

		if err := rows.Scan(
			&item.OrderID,
			&item.TotalPrice,
			&item.OrderDate,
			&item.PaymentStatus,
			&paidAt,
		); err != nil {
			return nil, err
		}

		item.PaidAt = paidAt
		result = append(result, item)
	}

	return result, nil
}
