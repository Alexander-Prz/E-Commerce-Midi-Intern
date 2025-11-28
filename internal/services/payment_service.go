package services

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"context"
	"errors"
)

// List available payment methods
func ListPaymentMethods(ctx context.Context) ([]repository.PaymentMethod, error) {
	return repository.GetPaymentMethods(ctx, db.Pool)
}

// StartPaymentForOrder creates a pending payment record for an order
func StartPaymentForOrder(ctx context.Context, orderID, methodID int, amount float64) (int, error) {
	// verify order exists and not deleted
	// You may want to add a repository function to verify order; for now we assume order exists
	return repository.CreatePayment(ctx, db.Pool, orderID, methodID, amount)
}

// ConfirmPayment marks payment as Paid and returns error if fails
func ConfirmPayment(ctx context.Context, paymentID int) error {
	// Check payment exists
	p, err := repository.GetPaymentByID(ctx, db.Pool, paymentID)
	if err != nil {
		return err
	}

	if p.PaymentStatus == "Paid" {
		return errors.New("payment already paid")
	}

	// Update status to Paid (and create log)
	if err := repository.UpdatePaymentStatus(ctx, db.Pool, paymentID, "Paid"); err != nil {
		return err
	}

	// Optionally you can perform additional post-payment actions here
	// e.g., send notification, grant license, etc.

	return nil
}

// FailPayment sets payment status to Failed
func FailPayment(ctx context.Context, paymentID int) error {
	_, err := repository.GetPaymentByID(ctx, db.Pool, paymentID)
	if err != nil {
		return err
	}
	return repository.UpdatePaymentStatus(ctx, db.Pool, paymentID, "Failed")
}

// GetPaymentsForOrder
func GetPaymentsForOrder(ctx context.Context, orderID int) ([]repository.Payment, error) {
	return repository.GetPaymentsByOrderID(ctx, db.Pool, orderID)
}

func GetAllTransactions(ctx context.Context) ([]repository.AdminTransaction, error) {
	return repository.GetAllTransactions(ctx, db.Pool)
}
