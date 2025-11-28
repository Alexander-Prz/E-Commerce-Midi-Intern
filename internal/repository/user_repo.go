package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserAuth struct {
	AuthID       int
	Email        string
	PasswordHash string
	Role         string
	Username     string
	CustomerID   int
	DeveloperID  int
}

type UserDetail struct {
	AuthID    int
	Email     string
	Role      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

type DeveloperDetail struct {
	DeveloperID   int
	DeveloperName string
	CreatedAt     time.Time
	DeletedAt     *time.Time

	AuthID      int
	Email       string
	AuthDeleted *time.Time
}

type AccountDetail struct {
	AuthID        int
	Email         string
	Role          string
	CreatedAt     time.Time
	DeletedAt     *time.Time
	DeveloperID   *int       // nil if not a developer
	DeveloperName *string    // nil if not a developer
	AuthDeleted   *time.Time // for devs linked auth deletion
}

func GetUserAuthByEmail(ctx context.Context, db *pgxpool.Pool, email string) (*UserAuth, error) {
	query := `
        SELECT authid, email, passwordhash, role
        FROM userauth
        WHERE email = $1
			AND deleted_at IS NULL
        LIMIT 1;
    `

	row := db.QueryRow(ctx, query, email)

	var ua UserAuth
	err := row.Scan(&ua.AuthID, &ua.Email, &ua.PasswordHash, &ua.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if ua.Role == "admin" {
		ua.Username = "Administrator"
		ua.CustomerID = 0
		return &ua, nil
	}

	if ua.Role == "developer" {
		// fetch developer info by authid
		devID, devName, err := GetDeveloperByAuthID(ctx, db, ua.AuthID)
		if err != nil {
			return nil, err
		}
		ua.DeveloperID = devID
		ua.Username = devName
		// we don't set CustomerID for developers
		ua.CustomerID = 0
		// optionally store devID somewhere; add field if you want (e.g., DeveloperID int)
		return &ua, nil
	}

	// USER → fetch username + customerID in one query
	username, customerID, err := GetCustomerInfoByAuthID(ctx, db, ua.AuthID)
	if err != nil {
		return nil, err
	}

	ua.Username = username
	ua.CustomerID = customerID
	return &ua, nil
}

func GetCustomerInfoByAuthID(ctx context.Context, db *pgxpool.Pool, authID int) (string, int, error) {
	var username string
	var customerID int

	query := `
        SELECT username, customerid
        FROM customers
        WHERE authid = $1;
    `

	err := db.QueryRow(ctx, query, authID).Scan(&username, &customerID)
	if err != nil {
		return "", 0, err
	}

	return username, customerID, nil
}

func RegisterUser(ctx context.Context, tx pgx.Tx, email, passwordHash, username string) error {

	var authID int
	queryUser := `
        INSERT INTO userauth (email, passwordhash, role)
        VALUES ($1, $2, 'user')
        RETURNING authid;
    `
	err := tx.QueryRow(ctx, queryUser, email, passwordHash).Scan(&authID)
	if err != nil {
		return err
	}

	// Insert Customers
	queryCustomer := `
        INSERT INTO customers (username, email, authid)
        VALUES ($1, $2, $3);
    `
	_, err = tx.Exec(ctx, queryCustomer, username, email, authID)
	if err != nil {
		return err
	}

	return nil
}

func RegisterAdmin(ctx context.Context, tx pgx.Tx, email, passwordHash string) error {

	// Insert UserAuth
	var authID int
	queryUser := `
        INSERT INTO userauth (email, passwordhash, role)
        VALUES ($1, $2, 'admin')
        RETURNING authid;
    `
	err := tx.QueryRow(ctx, queryUser, email, passwordHash).Scan(&authID)
	if err != nil {
		return err
	}

	return nil
}

func GetAllUsers(ctx context.Context, db *pgxpool.Pool) ([]UserDetail, error) {
	query := `
        SELECT authid, email, role, created_at, deleted_at
        FROM userauth
        WHERE role = 'user'
        ORDER BY authid;
    `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UserDetail
	for rows.Next() {
		var u UserDetail
		if err := rows.Scan(
			&u.AuthID,
			&u.Email,
			&u.Role,
			&u.CreatedAt,
			&u.DeletedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}

func GetUserByID(ctx context.Context, db *pgxpool.Pool, authID int) (*UserDetail, error) {
	var u UserDetail

	err := db.QueryRow(ctx,
		`SELECT authid, email, role, created_at, deleted_at
		 FROM userauth
		 WHERE authid = $1`,
		authID,
	).Scan(&u.AuthID, &u.Email, &u.Role, &u.CreatedAt, &u.DeletedAt)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func SoftDeleteUser(ctx context.Context, db *pgxpool.Pool, authID int) error {
	_, err := db.Exec(ctx,
		`UPDATE userauth
		 SET deleted_at = NOW()
		 WHERE authid = $1
		   AND role = 'user'
		   AND deleted_at IS NULL`,
		authID,
	)
	return err
}

func RestoreUser(ctx context.Context, db *pgxpool.Pool, authID int) error {
	_, err := db.Exec(ctx,
		`UPDATE userauth
		 SET deleted_at = NULL
		 WHERE authid = $1
		   AND role = 'user'
		   AND deleted_at IS NOT NULL`,
		authID,
	)
	return err
}

func RegisterDeveloper(ctx context.Context, tx pgx.Tx, email, passwordHash, devName string) error {
	var authID int
	queryUser := `
        INSERT INTO userauth (email, passwordhash, role)
        VALUES ($1, $2, 'developer')
        RETURNING authid;
    `
	if err := tx.QueryRow(ctx, queryUser, email, passwordHash).Scan(&authID); err != nil {
		return err
	}

	queryDev := `
        INSERT INTO developers (developername, authid)
        VALUES ($1, $2);
    `
	if _, err := tx.Exec(ctx, queryDev, devName, authID); err != nil {
		return err
	}

	return nil
}

func GetAllDevelopers(ctx context.Context, db *pgxpool.Pool) ([]DeveloperDetail, error) {
	query := `
        SELECT d.developerid, d.developername, d.created_at, d.deleted_at,
               ua.authid, ua.email, ua.deleted_at
        FROM developers d
        JOIN userauth ua ON ua.authid = d.authid
        WHERE ua.role = 'developer'
        ORDER BY d.developerid;
    `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DeveloperDetail
	for rows.Next() {
		var d DeveloperDetail
		if err := rows.Scan(
			&d.DeveloperID,
			&d.DeveloperName,
			&d.CreatedAt,
			&d.DeletedAt,
			&d.AuthID,
			&d.Email,
			&d.AuthDeleted,
		); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func GetDeveloperByAuthID(ctx context.Context, db *pgxpool.Pool, authID int) (int, string, error) {
	var devID int
	var name string
	query := `
        SELECT developerid, developername
        FROM developers
        WHERE authid = $1
        LIMIT 1;
    `
	if err := db.QueryRow(ctx, query, authID).Scan(&devID, &name); err != nil {
		if err == pgx.ErrNoRows {
			return 0, "", errors.New("developer not found")
		}
		return 0, "", err
	}
	return devID, name, nil
}

func GetDeveloperDetail(ctx context.Context, db *pgxpool.Pool, devID int) (*DeveloperDetail, error) {
	var d DeveloperDetail

	query := `
        SELECT d.developerid, d.developername, d.created_at, d.deleted_at,
               ua.authid, ua.email, ua.deleted_at
        FROM developers d
        JOIN userauth ua ON ua.authid = d.authid
        WHERE d.developerid = $1;
    `

	err := db.QueryRow(ctx, query, devID).Scan(
		&d.DeveloperID,
		&d.DeveloperName,
		&d.CreatedAt,
		&d.DeletedAt,
		&d.AuthID,
		&d.Email,
		&d.AuthDeleted,
	)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func GetAllAccounts(ctx context.Context, db *pgxpool.Pool) ([]AccountDetail, error) {
	query := `
        SELECT ua.authid, ua.email, ua.role, ua.created_at, ua.deleted_at,
               d.developerid, d.developername, d.deleted_at AS auth_deleted
        FROM userauth ua
        LEFT JOIN developers d ON ua.authid = d.authid
        ORDER BY ua.authid;
    `

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AccountDetail
	for rows.Next() {
		var a AccountDetail
		if err := rows.Scan(
			&a.AuthID,
			&a.Email,
			&a.Role,
			&a.CreatedAt,
			&a.DeletedAt,
			&a.DeveloperID,
			&a.DeveloperName,
			&a.AuthDeleted,
		); err != nil {
			return nil, err
		}
		list = append(list, a)
	}

	return list, nil
}

func GetAccountByAuthID(ctx context.Context, db *pgxpool.Pool, authID int) (*AccountDetail, error) {
	var a AccountDetail

	query := `
        SELECT ua.authid, ua.email, ua.role, ua.created_at, ua.deleted_at,
               d.developerid, d.developername, d.deleted_at AS auth_deleted
        FROM userauth ua
        LEFT JOIN developers d ON ua.authid = d.authid
        WHERE ua.authid = $1;
    `

	err := db.QueryRow(ctx, query, authID).Scan(
		&a.AuthID,
		&a.Email,
		&a.Role,
		&a.CreatedAt,
		&a.DeletedAt,
		&a.DeveloperID,
		&a.DeveloperName,
		&a.AuthDeleted,
	)

	if err != nil {
		return nil, err
	}

	return &a, nil
}
