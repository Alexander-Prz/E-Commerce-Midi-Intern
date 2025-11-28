package services

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"context"
)

func GetOrderHistory(ctx context.Context, customerID int) ([]repository.OrderHistoryItem, error) {
	return repository.GetOrderHistory(ctx, db.Pool, customerID)
}
