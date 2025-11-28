package services

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"context"
	"fmt"
	"strings"
)

func AllGames(ctx context.Context, page int) ([]repository.GameList, int, error) {
	const pageSize = 10

	allGames, err := repository.GetAllGames(ctx, db.Pool)
	if err != nil {
		return nil, 0, err
	}

	total := len(allGames)
	if total == 0 {
		return nil, 0, nil
	}

	totalPages := (total + pageSize - 1) / pageSize
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	return allGames[start:end], totalPages, nil
}

func GameDetails(id int) {
	ctx := context.Background()

	details, err := repository.GetGameDetails(ctx, db.Pool, id)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Title:", details.Title)
	fmt.Println("Price:", details.Price)
	fmt.Println("Developer:", details.DeveloperName)
	fmt.Println("Genres:", strings.Join(details.Genres, ", "))
	fmt.Printf("Year: %s\n", details.ReleaseDate.Format("2006-01-02"))

}

func GamePrice(ctx context.Context, gameID int) (float64, error) {
	return repository.GetGamePrice(ctx, db.Pool, gameID)
}

func AddGame(ctx context.Context, title string, price float64, releaseDate string, developerID int) (int, error) {
	return repository.AddGame(ctx, db.Pool, title, price, releaseDate, developerID)
}

func RemoveGame(ctx context.Context, gameID int, role string, devID int) error {
	return repository.RemoveGame(ctx, db.Pool, gameID, role, devID)
}

func EditGameDetails(ctx context.Context, id int, title string, price float64, releaseDate string, devID int) error {

	return repository.UpdateGameDetails(
		ctx,
		db.Pool,
		id,
		title,
		price,
		releaseDate,
		devID,
	)
}

func AddGenreToGame(ctx context.Context, gameID, genreID int) error {
	return repository.AddGenreToGame(ctx, db.Pool, gameID, genreID)
}

func UpdateGameGenres(ctx context.Context, gameID int, genreIDs []int) error {
	return repository.UpdateGameGenres(ctx, db.Pool, gameID, genreIDs)
}
