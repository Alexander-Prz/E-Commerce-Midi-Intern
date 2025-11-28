package services

import (
	"GamesProject/internal/db"
	"GamesProject/internal/repository"
	"context"
)

func GetAllUsers(ctx context.Context) ([]repository.UserDetail, error) {
	return repository.GetAllUsers(ctx, db.Pool)
}

func BanUser(ctx context.Context, authID int) error {
	return repository.SoftDeleteUser(ctx, db.Pool, authID)
}

func UnbanUser(ctx context.Context, authID int) error {
	return repository.RestoreUser(ctx, db.Pool, authID)
}

func GetUserByID(ctx context.Context, authID int) (*repository.UserDetail, error) {
	return repository.GetUserByID(ctx, db.Pool, authID)
}

func GetDeveloperDetail(ctx context.Context, devID int) (*repository.DeveloperDetail, error) {
	return repository.GetDeveloperDetail(ctx, db.Pool, devID)
}

func GetAllDevelopers(ctx context.Context) ([]repository.DeveloperDetail, error) {
	return repository.GetAllDevelopers(ctx, db.Pool)
}

func GetAllAccounts(ctx context.Context) ([]repository.AccountDetail, error) {
	return repository.GetAllAccounts(ctx, db.Pool)
}

func GetAccountByAuthID(ctx context.Context, authID int) (*repository.AccountDetail, error) {
	// Call the repository function
	account, err := repository.GetAccountByAuthID(ctx, db.Pool, authID)
	if err != nil {
		return nil, err
	}

	return account, nil
}
