package repository

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Developer struct {
	DeveloperID   int
	DeveloperName string
}

type GameSalesReport struct {
	GameID    int
	Title     string
	UnitsSold int
	Revenue   float64
}

func GetDeveloperByID(ctx context.Context, db *pgxpool.Pool, developerID int) (*Developer, error) {
	query := `
        SELECT developerid, developername
        FROM developers
        WHERE developerid = $1;
    `

	var d Developer

	err := db.QueryRow(ctx, query, developerID).Scan(
		&d.DeveloperID,
		&d.DeveloperName,
	)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func GetDeveloperGames(ctx context.Context, db *pgxpool.Pool, developerID int, page int, pageSize int) ([]GameList, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// count
	var total int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM games 
		 WHERE developerid = $1 AND deleted_at IS NULL`,
		developerID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// fetch page
	rows, err := db.Query(ctx,
		`SELECT gameid, title
		 FROM games
		 WHERE developerid = $1 AND deleted_at IS NULL
		 ORDER BY gameid
		 LIMIT $2 OFFSET $3`,
		developerID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []GameList
	for rows.Next() {
		var g GameList
		if err := rows.Scan(&g.GameID, &g.Title); err != nil {
			return nil, 0, err
		}
		list = append(list, g)
	}

	return list, totalPages, nil
}

func IsGameOwnedByDeveloper(ctx context.Context, db *pgxpool.Pool, devID, gameID int) (bool, error) {
	var count int
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) 
         FROM games 
         WHERE gameid=$1 AND developerid=$2 AND deleted_at IS NULL`,
		gameID, devID,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetDeveloperSalesReport(ctx context.Context, db *pgxpool.Pool, developerID int) ([]GameSalesReport, error) {

	query := `
		SELECT
			g.gameid,
			g.title,
			COALESCE(SUM(oi.quantity), 0) AS units_sold,
			COALESCE(SUM(oi.quantity * oi.priceatpurchase), 0) AS revenue
		FROM games g
		LEFT JOIN orderitems oi
			ON g.gameid = oi.gameid
			AND oi.deleted_at IS NULL
		WHERE g.developerid = $1
		  AND g.deleted_at IS NULL
		GROUP BY g.gameid, g.title
		ORDER BY revenue DESC;
	`

	rows, err := db.Query(ctx, query, developerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []GameSalesReport

	for rows.Next() {
		var r GameSalesReport
		if err := rows.Scan(&r.GameID, &r.Title, &r.UnitsSold, &r.Revenue); err != nil {
			return nil, err
		}
		list = append(list, r)
	}

	return list, rows.Err()
}
