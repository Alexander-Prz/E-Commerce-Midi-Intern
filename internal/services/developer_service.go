package services

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"GamesProject/internal/utils"
	"context"
	"fmt"
)

func DeveloperExists(ctx context.Context, id int) bool {
	_, err := repository.GetDeveloperByID(ctx, db.Pool, id)
	return err == nil
}

func GetValidDeveloperID(ctx context.Context, prompt string) int {
	for {
		id := utils.ReadInt(prompt)

		if DeveloperExists(ctx, id) {
			return id
		}

		fmt.Println("Invalid Developer ID! Please enter an existing ID.")
	}
}

func GetDeveloperByAuthID(ctx context.Context, authID int) (int, string, error) {
	return repository.GetDeveloperByAuthID(ctx, db.Pool, authID)
}

func DeveloperGames(ctx context.Context, developerID int, page int) ([]repository.GameList, int, error) {

	const pageSize = 10

	return repository.GetDeveloperGames(
		ctx,
		db.Pool,
		developerID,
		page,
		pageSize,
	)
}

func GameOwnedByDeveloper(ctx context.Context, devID, gameID int) (bool, error) {
	return repository.IsGameOwnedByDeveloper(ctx, db.Pool, devID, gameID)
}

func DeveloperSalesReport(ctx context.Context, devID int) ([]repository.GameSalesReport, error) {
	return repository.GetDeveloperSalesReport(ctx, db.Pool, devID)
}
