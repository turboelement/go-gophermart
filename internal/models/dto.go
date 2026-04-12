package models

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
