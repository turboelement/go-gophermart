package service

import (
	"context"

	"go-gophermart/internal/models"
	"go-gophermart/internal/repository"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (svc *OrderService) CreateOrder(ctx context.Context, orderNumber, userID string) (string, error) {
	order, err := svc.repo.CreateOrder(ctx, orderNumber, userID)
	if err != nil {
		return "", err
	}
	return order.ID, nil
}

func (svc *OrderService) GetOrdersByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	return svc.repo.GetOrdersByUserID(ctx, userID)
}
