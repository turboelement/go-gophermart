package repository

import (
	"context"
	"errors"
	"fmt"

	"go-gophermart/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPostgresRepository struct {
	db *pgxpool.Pool
}

func NewUserPostgresRepository(pool *pgxpool.Pool) *UserPostgresRepository {
	return &UserPostgresRepository{db: pool}
}

func (repo *UserPostgresRepository) CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error) {
	user := &models.User{
		Login:        login,
		PasswordHash: passwordHash,
	}

	err := repo.db.QueryRow(ctx,
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
		login,
		passwordHash,
	).Scan(&user.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrLoginAlreadyExists
		}
		return nil, fmt.Errorf("error saving to database: %w", err)
	}

	return user, nil
}

func (repo *UserPostgresRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}

	err := repo.db.QueryRow(ctx,
		`SELECT id, login, password_hash FROM users WHERE login = $1`,
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLoginNotFound
		}
		return nil, err
	}

	return user, nil
}
