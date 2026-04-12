package service

import (
	"context"
	"errors"

	"go-gophermart/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (svc *UserService) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", errors.New("login and password are required")
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return "", err
	}

	user, err := svc.repo.CreateUser(ctx, login, passwordHash)
	if err != nil {
		if errors.Is(err, repository.ErrLoginAlreadyExists) {
			return "", repository.ErrLoginAlreadyExists
		}
		return "", err
	}

	return user.ID, nil
}

func (svc *UserService) Login(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", errors.New("login and password are required")
	}

	user, err := svc.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrLoginNotFound) {
			return "", errors.New("login or password are incorrect")
		}
		return "", err
	}

	ok := checkPasswordHash(password, user.PasswordHash)
	if !ok {
		return "", errors.New("login or password are incorrect")
	}

	return user.ID, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
