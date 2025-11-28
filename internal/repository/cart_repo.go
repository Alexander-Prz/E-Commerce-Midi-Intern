package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CartItem struct {
	OrderItemID     int
	GameID          int
	Title           string
	Quantity        int
	PriceAtPurchase float64
}

type Cart struct {
	OrderID int
	Items   []CartItem
	Total   float64
}

/*
GetActiveCart – creates an empty cart if not exists
*/
func GetActiveCart(ctx context.Context, db *pgxpool.Pool, customerID int) (int, error) {
	var orderID int

	query := `
        SELECT orderid 
        FROM orders
        WHERE customerid = $1
          AND totalprice = 0
          AND deleted_at IS NULL
        LIMIT 1;
    `
	err := db.QueryRow(ctx, query, customerID).Scan(&orderID)
	if err == nil {
		return orderID, nil
	}

	// create new cart
	insert := `
        INSERT INTO orders (customerid, totalprice)
        VALUES ($1, 0)
        RETURNING orderid;
    `
	err = db.QueryRow(ctx, insert, customerID).Scan(&orderID)
	return orderID, err
}

/*
AddItemToCart
*/
func AddItemToCart(ctx context.Context, db *pgxpool.Pool, orderID, gameID, qty int, price float64) error {
	query := `
        INSERT INTO orderitems (orderid, gameid, quantity, priceatpurchase)
        VALUES ($1, $2, $3, $4);
    `
	_, err := db.Exec(ctx, query, orderID, gameID, qty, price)
	return err
}

/*
GetCartItems
*/
func GetCartItems(ctx context.Context, db *pgxpool.Pool, orderID int) ([]CartItem, float64, error) {
	query := `
        SELECT 
            oi.orderitemid,
            oi.gameid,
            g.title,
            oi.quantity,
            oi.priceatpurchase
        FROM orderitems oi
        JOIN games g ON g.gameid = oi.gameid
        WHERE oi.orderid = $1
          AND oi.deleted_at IS NULL;
    `

	rows, err := db.Query(ctx, query, orderID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []CartItem{}
	total := float64(0)

	for rows.Next() {
		var ci CartItem
		if err := rows.Scan(&ci.OrderItemID, &ci.GameID, &ci.Title, &ci.Quantity, &ci.PriceAtPurchase); err != nil {
			return nil, 0, err
		}
		items = append(items, ci)
		total += float64(ci.Quantity) * ci.PriceAtPurchase
	}

	return items, total, nil
}

/*
UpdateCartItemQty
*/
func UpdateCartItemQty(ctx context.Context, db *pgxpool.Pool, orderItemID, qty int) error {
	query := `
        UPDATE orderitems
        SET quantity = $1
        WHERE orderitemid = $2
          AND deleted_at IS NULL;
    `
	_, err := db.Exec(ctx, query, qty, orderItemID)
	return err
}

/*
RemoveCartItem
*/
func RemoveCartItem(ctx context.Context, db *pgxpool.Pool, orderItemID int) error {
	query := `
        UPDATE orderitems
        SET deleted_at = NOW()
        WHERE orderitemid = $1;
    `
	_, err := db.Exec(ctx, query, orderItemID)
	return err
}

/*
ClearCart
*/
func ClearCart(ctx context.Context, db *pgxpool.Pool, orderID int) error {
	query := `
        UPDATE orderitems
        SET deleted_at = NOW()
        WHERE orderid = $1;
    `
	_, err := db.Exec(ctx, query, orderID)
	return err
}

/*
Checkout – finalizes the order and sets total price
*/
func Checkout(ctx context.Context, db *pgxpool.Pool, orderID int, total float64) error {
	query := `
        UPDATE orders
        SET totalprice = $1
        WHERE orderid = $2
          AND deleted_at IS NULL;
    `
	_, err := db.Exec(ctx, query, total, orderID)
	return err
}
