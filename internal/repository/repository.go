package repository

import (
	"context"
	"errors"

	"go-gophermart/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, passwordHash string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, orderNumber, userID string) (*models.Order, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID string) ([]models.Order, error)
	GetOrdersForAccrual(ctx context.Context, count int) ([]models.Order, error)
	UpdateOrderByNumber(ctx context.Context, orderNumber string, status models.OrderStatus, accrual float64) error
}

type BalanceRepository interface {
	GetBalanceByUserID(ctx context.Context, userID string) (*models.UserBalance, error)
	Withdraw(ctx context.Context, userID string, orderNumber string, sum float64) error
	GetWithdrawalsByUserID(ctx context.Context, userID string) ([]models.Balance, error)
}

var (
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrLoginNotFound      = errors.New("login not found")

	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderDuplicate     = errors.New("order duplicate by user")
	ErrOrderNotFound      = errors.New("order not found")

	ErrBalanceInsufficientFunds = errors.New("insufficient funds")
)
