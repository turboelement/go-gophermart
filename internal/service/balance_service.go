package service

import (
	"context"

	"go-gophermart/internal/models"
	"go-gophermart/internal/repository"
)

type BalanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

func (svc *BalanceService) GetBalanceByUserID(ctx context.Context, userID string) (*models.UserBalance, error) {
	return svc.repo.GetBalanceByUserID(ctx, userID)
}

func (svc *BalanceService) Withdraw(ctx context.Context, userID string, orderNumber string, sum float64) error {
	return svc.repo.Withdraw(ctx, userID, orderNumber, sum)
}

func (svc *BalanceService) GetWithdrawalsByUserID(ctx context.Context, userID string) ([]models.Balance, error) {
	return svc.repo.GetWithdrawalsByUserID(ctx, userID)
}
