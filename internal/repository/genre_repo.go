package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GenreList struct {
	GenreID   int
	GenreName string
}

func GetAllGenre(ctx context.Context, db *pgxpool.Pool) ([]GenreList, error) {
	query := `
        SELECT genreid, genrename
        FROM genres
        WHERE deleted_at IS NULL
        ORDER BY genreid;
    `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genre []GenreList

	for rows.Next() {
		var g GenreList
		if err := rows.Scan(&g.GenreID, &g.GenreName); err != nil {
			return nil, err
		}
		genre = append(genre, g)
	}

	return genre, nil
}

func AddGenre(ctx context.Context, db *pgxpool.Pool, name string) error {
	query := `
        INSERT INTO genres (genrename)
        VALUES ($1);
    `

	_, err := db.Exec(ctx, query, name)
	return err
}

func RemoveGenre(ctx context.Context, db *pgxpool.Pool, genreID int) error {
	query := `
        UPDATE genres
        SET deleted_at = NOW()
        WHERE genreid = $1 AND deleted_at IS NULL;
    `

	_, err := db.Exec(ctx, query, genreID)
	return err
}
