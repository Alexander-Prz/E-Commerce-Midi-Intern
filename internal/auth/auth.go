package auth

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var CurrentUser *repository.UserAuth = nil

func Login(ctx context.Context, email, password string) error {
	user, err := repository.GetUserAuthByEmail(ctx, db.Pool, email)
	if err != nil {
		return errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return errors.New("invalid email or password")
	}

	CurrentUser = user
	return nil
}

func Logout() {
	CurrentUser = nil
}

func Register(ctx context.Context, email, password, username string) error {

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Perform the registration inside the transaction
	err = repository.RegisterUser(ctx, tx, email, string(hash), username)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}

func RegisterForAdmin(ctx context.Context, email, password string) error {

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Perform the registration inside the transaction
	err = repository.RegisterAdmin(ctx, tx, email, string(hash))
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}

func RegisterForDeveloper(ctx context.Context, email, password, devName string) error {

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Perform the registration inside the transaction
	err = repository.RegisterDeveloper(ctx, tx, email, string(hash), devName)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}
