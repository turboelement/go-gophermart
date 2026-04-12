package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"go-gophermart/internal/middleware"
	"go-gophermart/internal/repository"
	"go-gophermart/internal/service"
	"go-gophermart/internal/utils"

	"go.uber.org/zap"
)

type OrderHandler struct {
	svc    *service.OrderService
	logger *zap.Logger
}

func NewOrderHandler(svc *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		h.logger.Debug("Failed to get user ID from context", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Debug("Cannot read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	order := strings.TrimSpace(string(body))
	if order == "" {
		h.logger.Debug("Number can not be empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !utils.IsValidLuhn(order) {
		h.logger.Debug("Number is not valid")
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	_, err = h.svc.CreateOrder(r.Context(), order, userID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderDuplicate) {
			h.logger.Debug("Order duplicated by user", zap.Error(err))
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, repository.ErrOrderAlreadyExists) {
			h.logger.Debug("Order already exists", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		h.logger.Error("Failed to create order", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		h.logger.Debug("Failed to get user ID from context", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	orders, err := h.svc.GetOrdersByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		h.logger.Error("Failed to encode orders", zap.Error(err))
	}
}
