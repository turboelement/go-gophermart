package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-gophermart/internal/auth"
	"go-gophermart/internal/models"
	"go-gophermart/internal/repository"
	"go-gophermart/internal/service"

	"go.uber.org/zap"
)

type UserHandler struct {
	svc    *service.UserService
	key    string
	secure bool
	logger *zap.Logger
}

func NewUserHandler(svc *service.UserService, key string, secure bool, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		svc:    svc,
		key:    key,
		secure: secure,
		logger: logger,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("Invalid JSON", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Login == "" || req.Password == "" {
		h.logger.Debug("Login and Password are required")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := h.svc.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrLoginAlreadyExists) {
			h.logger.Debug("Login already exists", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		h.logger.Error("Failed to register user", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := auth.SetAuthCookie(w, userID, h.key, h.secure); err != nil {
		h.logger.Error("Failed to set auth cookie", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("Invalid JSON", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Login == "" || req.Password == "" {
		h.logger.Debug("Login and Password are required")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := h.svc.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Debug("Failed to login user", zap.String("login", req.Login), zap.Error(err))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err := auth.SetAuthCookie(w, userID, h.key, h.secure); err != nil {
		h.logger.Error("Failed to set auth cookie", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
