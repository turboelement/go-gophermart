package repository

import (
	"context"
	"fmt"

	"go-gophermart/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BalancePostgresRepository struct {
	db *pgxpool.Pool
}

func NewBalanceRepository(pool *pgxpool.Pool) *BalancePostgresRepository {
	return &BalancePostgresRepository{db: pool}
}

func (repo *BalancePostgresRepository) GetBalanceByUserID(ctx context.Context, userID string) (*models.UserBalance, error) {
	query := `
		SELECT 
			COALESCE((SELECT SUM(accrual) FROM orders WHERE user_id = $1 AND status = $2), 0) 
			- COALESCE((SELECT SUM(sum) FROM balances WHERE user_id = $1), 0) AS current,
			
			COALESCE((SELECT SUM(sum) FROM balances WHERE user_id = $1), 0) AS withdrawn;
	`
	userBalance := &models.UserBalance{}
	err := repo.db.QueryRow(ctx,
		query,
		userID,
		models.OrderStatusProcessed,
	).Scan(&userBalance.Current, &userBalance.Withdrawn)

	if err != nil {
		return nil, err
	}

	return userBalance, nil
}

func (repo *BalancePostgresRepository) Withdraw(ctx context.Context, userID string, orderNumber string, sum float64) error {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userBalance, err := repo.GetBalanceByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if userBalance.Current < sum {
		return ErrBalanceInsufficientFunds
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO balances (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, NOW())",
		userID,
		orderNumber,
		sum,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (repo *BalancePostgresRepository) GetWithdrawalsByUserID(ctx context.Context, userID string) ([]models.Balance, error) {
	rows, err := repo.db.Query(ctx,
		"SELECT id, user_id, order_number, sum, processed_at FROM balances WHERE user_id = $1 ORDER BY processed_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying user withdrawals: %w", err)
	}
	defer rows.Close()

	withdrawals := []models.Balance{}
	for rows.Next() {
		var balance models.Balance
		if err := rows.Scan(&balance.ID, &balance.UserID, &balance.OrderNumber, &balance.Sum, &balance.ProcessedAt); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		withdrawals = append(withdrawals, balance)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return withdrawals, nil
}
