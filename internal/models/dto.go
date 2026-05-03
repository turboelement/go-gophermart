package models

import "time"

type UserRegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type WithdrawByOrderNumberRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type AccrualResponse struct {
	AccrualOrder  *AccrualOrder
	StatusCode    int
	RetryInterval time.Duration
}
