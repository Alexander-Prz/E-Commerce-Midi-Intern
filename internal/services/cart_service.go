package services

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"context"
	"fmt"
)

func AddToCart(ctx context.Context, customerID, gameID, qty int, price float64) error {
	orderID, err := repository.GetActiveCart(ctx, db.Pool, customerID)
	if err != nil {
		return err
	}

	return repository.AddItemToCart(ctx, db.Pool, orderID, gameID, qty, price)
}

func ViewCart(ctx context.Context, customerID int) (*repository.Cart, error) {
	orderID, err := repository.GetActiveCart(ctx, db.Pool, customerID)
	if err != nil {
		return nil, err
	}

	items, total, err := repository.GetCartItems(ctx, db.Pool, orderID)
	if err != nil {
		return nil, err
	}

	return &repository.Cart{
		OrderID: orderID,
		Items:   items,
		Total:   total,
	}, nil
}

func UpdateQuantity(ctx context.Context, orderItemID, qty int) error {
	return repository.UpdateCartItemQty(ctx, db.Pool, orderItemID, qty)
}

func RemoveFromCart(ctx context.Context, orderItemID int) error {
	return repository.RemoveCartItem(ctx, db.Pool, orderItemID)
}

func ClearCart(ctx context.Context, customerID int) error {
	orderID, err := repository.GetActiveCart(ctx, db.Pool, customerID)
	if err != nil {
		return err
	}
	return repository.ClearCart(ctx, db.Pool, orderID)
}

func CheckoutCart(ctx context.Context, customerID int) (int, float64, error) {
	orderID, err := repository.GetActiveCart(ctx, db.Pool, customerID)
	if err != nil {
		return 0, 0, err
	}

	items, total, err := repository.GetCartItems(ctx, db.Pool, orderID)
	if err != nil {
		return 0, 0, err
	}

	if len(items) == 0 {
		return 0, 0, fmt.Errorf("cart is empty")
	}

	// finalize order by writing total price
	if err := repository.Checkout(ctx, db.Pool, orderID, total); err != nil {
		return 0, 0, err
	}

	return orderID, total, nil
}
