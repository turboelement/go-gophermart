package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-gophermart/internal/middleware"
	"go-gophermart/internal/models"
	"go-gophermart/internal/repository"
	"go-gophermart/internal/service"
	"go-gophermart/internal/utils"

	"go.uber.org/zap"
)

type BalanceHandler struct {
	svc    *service.BalanceService
	logger *zap.Logger
}

func NewBalanceHandler(svc *service.BalanceService, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *BalanceHandler) GetBalanceByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		h.logger.Debug("Failed to get user ID from context", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	userBalance, err := h.svc.GetBalanceByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userBalance); err != nil {
		h.logger.Error("Failed to encode balance", zap.Error(err))
	}
}

func (h *BalanceHandler) WithdrawByOrderNumber(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		h.logger.Debug("Failed to get user ID from context", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var req models.WithdrawByOrderNumberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("Invalid JSON", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Order == "" || req.Sum <= 0 {
		h.logger.Debug("Incorrect order or sum", zap.String("order", req.Order), zap.Float64("sum", req.Sum))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !utils.IsValidLuhn(req.Order) {
		h.logger.Debug("Number is not valid")
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	err = h.svc.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, repository.ErrBalanceInsufficientFunds) {
			h.logger.Debug("Insufficent funds", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
			return
		}
		h.logger.Error("Failed to withdraw", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) WithdrawalsByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		h.logger.Debug("Failed to get user ID from context", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.svc.GetWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get withdrawals", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Error("Failed to encode withdrawals", zap.Error(err))
	}
}
