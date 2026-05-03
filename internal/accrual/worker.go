package accrual

import (
	"context"
	"errors"
	"sync"
	"time"

	"go-gophermart/internal/models"
	"go-gophermart/internal/service"

	"go.uber.org/zap"
)

const maxOrdersCountForAccrual int = 10

type Worker struct {
	orderService  *service.OrderService
	accrualClient *Client
	logger        *zap.Logger
	interval      time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func NewWorker(orderService *service.OrderService, accrualClient *Client, logger *zap.Logger) *Worker {
	return &Worker{
		orderService:  orderService,
		accrualClient: accrualClient,
		logger:        logger,
		interval:      10 * time.Second,
		stopCh:        make(chan struct{}),
	}
}

func (aw *Worker) Start(ctx context.Context) {
	aw.wg.Add(1)
	go aw.run(ctx)
}

func (aw *Worker) Stop() {
	close(aw.stopCh)
	aw.wg.Wait()
}

func (aw *Worker) run(ctx context.Context) {
	defer aw.wg.Done()

	ticker := time.NewTicker(aw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			aw.logger.Info("Accrual worker stopped by context")
			return
		case <-aw.stopCh:
			aw.logger.Info("Accrual worker stopped")
			return
		case <-ticker.C:
			if err := aw.processOrders(ctx); err != nil {
				aw.logger.Error("Failed to process orders", zap.Error(err))
			}
		}
	}
}

func (aw *Worker) processOrders(ctx context.Context) error {
	orders, err := aw.orderService.GetOrdersForAccrual(ctx, maxOrdersCountForAccrual)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		return nil
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-aw.stopCh:
			return nil
		default:
			err := aw.processOrder(ctx, order)
			if err != nil {
				aw.logger.Error("Failed to process order",
					zap.String("number", order.Number),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

func (aw *Worker) processOrder(ctx context.Context, order models.Order) error {
	accrualResp, err := aw.accrualClient.GetAccrualOrderByNumber(ctx, order.Number)
	if err != nil {
		if errors.Is(err, ErrAccrualTooManyRequests) {
			aw.logger.Debug("Too many requests to accrual system",
				zap.Duration("retry after seconds", accrualResp.RetryInterval),
			)
			aw.interval = accrualResp.RetryInterval
			return nil
		}

		if errors.Is(err, ErrAccrualOrderNotRegistered) {
			aw.logger.Debug("Order not registered",
				zap.String("number", order.Number),
			)
			return nil
		}

		return err
	}

	switch accrualResp.AccrualOrder.Status {
	case models.OrderStatusRegistered:
		if order.Status == models.OrderStatusNew {
			err := aw.orderService.UpdateOrderByNumber(ctx, order.Number, models.OrderStatusProcessing, 0)
			if err != nil {
				return err
			}
			aw.logger.Info("Order processed",
				zap.String("number", order.Number),
				zap.String("status", string(models.OrderStatusProcessing)),
			)
		}

	case models.OrderStatusProcessing:
		if order.Status != models.OrderStatusProcessing {
			err := aw.orderService.UpdateOrderByNumber(ctx, order.Number, models.OrderStatusProcessing, 0)
			if err != nil {
				return err
			}
			aw.logger.Info("Order processed",
				zap.String("number", order.Number),
				zap.String("status", string(models.OrderStatusProcessing)),
			)
		}

	case models.OrderStatusInvalid:
		if err := aw.orderService.UpdateOrderByNumber(ctx, order.Number, models.OrderStatusInvalid, 0); err != nil {
			return err
		}
		aw.logger.Info("Order processed as invalid",
			zap.String("number", order.Number),
		)

	case models.OrderStatusProcessed:
		if err := aw.orderService.UpdateOrderByNumber(ctx, order.Number, models.OrderStatusProcessed, accrualResp.AccrualOrder.Accrual); err != nil {
			return err
		}
		aw.logger.Info("Order processed successfully",
			zap.String("number", order.Number),
			zap.Float64("accrual", accrualResp.AccrualOrder.Accrual),
		)
	}

	return nil
}
