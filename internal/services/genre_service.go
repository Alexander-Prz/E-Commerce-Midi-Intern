package services

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"context"
)

func AllGenres(ctx context.Context, page int) ([]repository.GenreList, int, error) {
	const pageSize = 10

	allGenres, err := repository.GetAllGenre(ctx, db.Pool)
	if err != nil {
		return nil, 0, err
	}

	total := len(allGenres)
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

	return allGenres[start:end], totalPages, nil
}

func AddGenre(ctx context.Context, name string) error {
	return repository.AddGenre(ctx, db.Pool, name)
}

func RemoveGenre(ctx context.Context, genreID int) error {
	return repository.RemoveGenre(ctx, db.Pool, genreID)
}
