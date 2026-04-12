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

type OrderPostgresRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderPostgresRepository {
	return &OrderPostgresRepository{db: pool}
}

func (repo *OrderPostgresRepository) CreateOrder(ctx context.Context, orderNumber, userID string) (*models.Order, error) {
	order := &models.Order{
		Number: orderNumber,
		UserID: userID,
		Status: models.OrderStatusNew,
	}

	err := repo.db.QueryRow(ctx,
		"INSERT INTO orders (number, user_id, status) VALUES ($1, $2, $3) RETURNING id, uploaded_at",
		orderNumber,
		userID,
		models.OrderStatusNew,
	).Scan(&order.ID, &order.UploadedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			oldOrder, err := repo.GetOrderByNumber(ctx, orderNumber)
			if err != nil {
				return nil, err
			}

			if oldOrder.UserID == userID {
				return oldOrder, ErrOrderDuplicate
			}
			return nil, ErrOrderAlreadyExists
		}
		return nil, fmt.Errorf("error saving to database: %w", err)
	}

	return order, nil
}

func (repo *OrderPostgresRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	order := &models.Order{}

	err := repo.db.QueryRow(ctx,
		"SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE number = $1",
		orderNumber,
	).Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

func (repo *OrderPostgresRepository) GetOrdersByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	rows, err := repo.db.Query(ctx,
		"SELECT id, number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying user orders: %w", err)
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return orders, nil
}
