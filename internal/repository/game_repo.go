package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameList struct {
	GameID int
	Title  string
}

type GameDetails struct {
	GameID        int
	Title         string
	Price         float64
	ReleaseDate   *time.Time
	DeveloperName string
	Genres        []string
}

func GetAllGames(ctx context.Context, db *pgxpool.Pool) ([]GameList, error) {
	query := `
        SELECT gameid, title
        FROM games
        WHERE deleted_at IS NULL
        ORDER BY gameid;
    `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []GameList

	for rows.Next() {
		var g GameList
		if err := rows.Scan(&g.GameID, &g.Title); err != nil {
			return nil, err
		}
		games = append(games, g)
	}

	return games, nil
}

func GetGameDetails(ctx context.Context, db *pgxpool.Pool, gameID int) (*GameDetails, error) {
	query := `
        SELECT 
            g.gameid,
            g.title,
            g.price,
            g.releasedate,
            d.developername
        FROM games g
        JOIN developers d ON d.developerid = g.developerid
        WHERE g.gameid = $1
          AND g.deleted_at IS NULL;
    `

	var gd GameDetails

	err := db.QueryRow(ctx, query, gameID).Scan(
		&gd.GameID,
		&gd.Title,
		&gd.Price,
		&gd.ReleaseDate,
		&gd.DeveloperName,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("game not found")
		}
		return nil, err
	}

	// Fetch Genres
	genres, err := GetGameGenres(ctx, db, gameID)
	if err != nil {
		return nil, err
	}

	gd.Genres = genres

	return &gd, nil
}

func GetGameGenres(ctx context.Context, db *pgxpool.Pool, gameID int) ([]string, error) {
	query := `
        SELECT ge.genrename
        FROM gamegenres gg
        JOIN genres ge ON ge.genreid = gg.genreid
        WHERE gg.gameid = $1
          AND ge.deleted_at IS NULL;
    `

	rows, err := db.Query(ctx, query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []string

	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}

	return genres, nil
}

func GetGamePrice(ctx context.Context, db *pgxpool.Pool, gameID int) (float64, error) {
	query := `
        SELECT price
        FROM games
        WHERE gameid = $1
          AND deleted_at IS NULL;
    `

	var price float64
	err := db.QueryRow(ctx, query, gameID).Scan(&price)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, errors.New("game not found")
		}
		return 0, err
	}

	return price, nil
}

func AddGame(ctx context.Context, db *pgxpool.Pool, title string, price float64, releaseDate string, developerID int) (int, error) {

	// Check developer
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM developers 
			WHERE developerid=$1 AND deleted_at IS NULL
		)`, developerID,
	).Scan(&exists)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, errors.New("developer not found")
	}

	// Insert and return ID
	var id int
	err = db.QueryRow(ctx,
		`INSERT INTO games (title, price, releasedate, developerid)
		 VALUES ($1, $2, $3, $4)
		 RETURNING gameid`,
		title, price, releaseDate, developerID,
	).Scan(&id)

	return id, err
}

func RemoveGame(ctx context.Context, db *pgxpool.Pool, gameID int, requesterRole string, requesterDevID int) error {

	// If requester is developer → check ownership
	if requesterRole == "developer" {
		var ownerID int
		err := db.QueryRow(ctx,
			`SELECT developerid FROM games WHERE gameid = $1 AND deleted_at IS NULL`,
			gameID,
		).Scan(&ownerID)

		if err != nil {
			return err
		}
		if ownerID != requesterDevID {
			return errors.New("you are not allowed to delete this game")
		}
	}

	// Admin OR owner developer → proceed delete
	_, err := db.Exec(ctx,
		`UPDATE games 
		 SET deleted_at = NOW() 
		 WHERE gameid = $1 AND deleted_at IS NULL`,
		gameID,
	)
	return err
}

func UpdateGameDetails(ctx context.Context, db *pgxpool.Pool, id int, title string, price float64, releaseDate string, devID int) error {

	// Step 1 — check ownership
	var ownerID int
	err := db.QueryRow(ctx,
		`SELECT developerid 
		 FROM games 
		 WHERE gameid=$1 AND deleted_at IS NULL`,
		id,
	).Scan(&ownerID)
	if err != nil {
		return err
	}

	// Step 2 — permission: only owner dev can edit
	if ownerID != devID {
		return errors.New("permission denied: you can only edit your own games")
	}

	// Step 3 — update allowed fields
	_, err = db.Exec(ctx,
		`UPDATE games
		 SET title=$1,
		     price=$2,
		     releasedate=$3
		 WHERE gameid=$4 
		   AND deleted_at IS NULL`,
		title, price, releaseDate, id,
	)

	return err
}

func AddGenreToGame(ctx context.Context, db *pgxpool.Pool, gameID, genreID int) error {
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM genres 
			WHERE genreid=$1 AND deleted_at IS NULL
		)`, genreID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("genre not found")
	}

	_, err = db.Exec(ctx,
		`INSERT INTO gamegenres (gameid, genreid)
		 VALUES ($1, $2)
		 ON CONFLICT (gameid, genreid) DO NOTHING`,
		gameID, genreID,
	)
	return err
}

func ClearGenresForGame(ctx context.Context, db *pgxpool.Pool, gameID int) error {
	_, err := db.Exec(ctx,
		`DELETE FROM gamegenres WHERE gameid=$1`,
		gameID,
	)
	return err
}

func UpdateGameGenres(ctx context.Context, db *pgxpool.Pool, gameID int, genreIDs []int) error {

	// Step 1: Clear old genres
	if err := ClearGenresForGame(ctx, db, gameID); err != nil {
		return err
	}

	// Step 2: Insert new ones
	for _, gid := range genreIDs {
		if err := AddGenreToGame(ctx, db, gameID, gid); err != nil {
			return err
		}
	}
	return nil
}
