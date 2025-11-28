package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentMethod struct {
	PaymentMethodID int
	Name            string
}

type Payment struct {
	PaymentID       int
	OrderID         int
	PaymentMethodID int
	AmountPaid      float64
	PaymentStatus   string
	CreatedAt       time.Time
	PaidAt          *time.Time
}

type PaymentLog struct {
	LogID     int
	PaymentID int
	OldStatus string
	NewStatus string
	ChangedAt time.Time
}

type AdminTransaction struct {
	OrderID       int
	CustomerID    int
	TotalPrice    float64
	OrderDate     time.Time
	PaymentStatus string
	PaidAt        *time.Time
}

// GetPaymentMethods returns available payment methods
func GetPaymentMethods(ctx context.Context, db *pgxpool.Pool) ([]PaymentMethod, error) {
	query := `
        SELECT paymentmethodid, name
        FROM paymentmethods
        ORDER BY paymentmethodid;
    `
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	methods := []PaymentMethod{}
	for rows.Next() {
		var m PaymentMethod
		if err := rows.Scan(&m.PaymentMethodID, &m.Name); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}

// CreatePayment inserts a pending payment and returns payment id
func CreatePayment(ctx context.Context, db *pgxpool.Pool, orderID, methodID int, amount float64) (int, error) {
	query := `
        INSERT INTO payments (orderid, paymentmethodid, amountpaid, paymentstatus)
        VALUES ($1, $2, $3, 'Pending')
        RETURNING paymentid;
    `
	var pid int
	err := db.QueryRow(ctx, query, orderID, methodID, amount).Scan(&pid)
	return pid, err
}

// GetPaymentByID
func GetPaymentByID(ctx context.Context, db *pgxpool.Pool, paymentID int) (*Payment, error) {
	query := `
        SELECT paymentid, orderid, paymentmethodid, amountpaid, paymentstatus, createdat, paidat
        FROM payments
        WHERE paymentid = $1;
    `
	var p Payment
	var paidAt *time.Time
	err := db.QueryRow(ctx, query, paymentID).Scan(
		&p.PaymentID,
		&p.OrderID,
		&p.PaymentMethodID,
		&p.AmountPaid,
		&p.PaymentStatus,
		&p.CreatedAt,
		&paidAt,
	)
	if err != nil {
		return nil, err
	}
	p.PaidAt = paidAt
	return &p, nil
}

// GetPaymentsByOrderID (returns all payments for an order)
func GetPaymentsByOrderID(ctx context.Context, db *pgxpool.Pool, orderID int) ([]Payment, error) {
	query := `
        SELECT paymentid, orderid, paymentmethodid, amountpaid, paymentstatus, createdat, paidat
        FROM payments
        WHERE orderid = $1
        ORDER BY createdat DESC;
    `
	rows, err := db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Payment{}
	for rows.Next() {
		var p Payment
		var paidAt *time.Time
		if err := rows.Scan(
			&p.PaymentID,
			&p.OrderID,
			&p.PaymentMethodID,
			&p.AmountPaid,
			&p.PaymentStatus,
			&p.CreatedAt,
			&paidAt,
		); err != nil {
			return nil, err
		}
		p.PaidAt = paidAt
		out = append(out, p)
	}
	return out, nil
}

// UpdatePaymentStatus updates status and optionally sets PaidAt when status = 'Paid'
func UpdatePaymentStatus(ctx context.Context, db *pgxpool.Pool, paymentID int, newStatus string) error {
	// Fetch current payment to know old status
	cur, err := GetPaymentByID(ctx, db, paymentID)
	if err != nil {
		return err
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update payment status and paidat when becoming Paid
	if newStatus == "Paid" {
		query := `
            UPDATE payments
            SET paymentstatus = $1, paidat = NOW()
            WHERE paymentid = $2;
        `
		if _, err := tx.Exec(ctx, query, newStatus, paymentID); err != nil {
			return err
		}
	} else {
		query := `
            UPDATE payments
            SET paymentstatus = $1
            WHERE paymentid = $2;
        `
		if _, err := tx.Exec(ctx, query, newStatus, paymentID); err != nil {
			return err
		}
	}

	// Insert log
	logQuery := `
        INSERT INTO paymentlogs (paymentid, oldstatus, newstatus)
        VALUES ($1, $2, $3);
    `
	if _, err := tx.Exec(ctx, logQuery, paymentID, cur.PaymentStatus, newStatus); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// GetPaymentMethodByID (helper)
func GetPaymentMethodByID(ctx context.Context, db *pgxpool.Pool, methodID int) (*PaymentMethod, error) {
	query := `
        SELECT paymentmethodid, name
        FROM paymentmethods
        WHERE paymentmethodid = $1;
    `
	var m PaymentMethod
	if err := db.QueryRow(ctx, query, methodID).Scan(&m.PaymentMethodID, &m.Name); err != nil {
		return nil, err
	}
	return &m, nil
}

func GetAllTransactions(ctx context.Context, db *pgxpool.Pool) ([]AdminTransaction, error) {
	query := `
        SELECT 
            o.orderid,
            o.customerid,
            o.totalprice,
            o.orderdate,
            p.paymentstatus,
            p.paidat
        FROM orders o
        JOIN payments p ON p.orderid = o.orderid
        WHERE o.deleted_at IS NULL
          AND p.paymentstatus = 'Paid'
        ORDER BY o.orderdate ASC;
    `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AdminTransaction
	for rows.Next() {
		var t AdminTransaction
		if err := rows.Scan(
			&t.OrderID,
			&t.CustomerID,
			&t.TotalPrice,
			&t.OrderDate,
			&t.PaymentStatus,
			&t.PaidAt,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	return list, nil
}
