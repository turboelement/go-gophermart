package router

import (
	"go-gophermart/internal/handler"
	"go-gophermart/internal/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type RouterDeps struct {
	UserHandler    *handler.UserHandler
	OrderHandler   *handler.OrderHandler
	BalanceHandler *handler.BalanceHandler
	CookieSecret   string
	Logger         *zap.Logger
}

func New(deps RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.AuthMiddleware(deps.CookieSecret, deps.Logger))

	// * `POST /api/user/register` — регистрация пользователя;
	r.Post("/api/user/register", deps.UserHandler.Register)

	// * `POST /api/user/login` — аутентификация пользователя;
	r.Post("/api/user/login", deps.UserHandler.Login)

	// * `POST /api/user/orders` — загрузка пользователем номера заказа для расчёта;
	r.Post("/api/user/orders", deps.OrderHandler.CreateOrder)

	// * `GET /api/user/orders` — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
	r.Get("/api/user/orders", deps.OrderHandler.GetOrders)

	// * `GET /api/user/balance` — получение текущего баланса счёта баллов лояльности пользователя;
	r.Get("/api/user/balance", deps.BalanceHandler.GetBalanceByUserID)

	// * `POST /api/user/balance/withdraw` — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
	r.Post("/api/user/balance/withdraw", deps.BalanceHandler.WithdrawByOrderNumber)

	// * `GET /api/user/withdrawals` — получение информации о выводе средств с накопительного счёта пользователем.
	r.Get("/api/user/withdrawals", deps.BalanceHandler.WithdrawalsByUserID)

	return r
}
