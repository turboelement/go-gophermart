package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-gophermart/internal/models"
)

var (
	ErrAccrualOrderNotRegistered = errors.New("accrual order not registered")
	ErrAccrualTooManyRequests    = errors.New("accrual too many requests")
	ErrAccrualInternalError      = errors.New("accrual internal error")
)

const defaultRetryInterval time.Duration = 5 * time.Second

type Client struct {
	url        string
	httpClient *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (ac *Client) GetAccrualOrderByNumber(ctx context.Context, orderNumber string) (*models.AccrualResponse, error) {
	accrualResp := &models.AccrualResponse{
		AccrualOrder:  nil,
		StatusCode:    http.StatusInternalServerError,
		RetryInterval: 0,
	}

	clientURL, err := url.JoinPath(ac.url, "/api/orders/", orderNumber)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, clientURL, nil)
	if err != nil {
		return accrualResp, err
	}

	resp, err := ac.httpClient.Do(r)
	if err != nil {
		return accrualResp, err
	}
	defer resp.Body.Close()

	accrualResp.StatusCode = resp.StatusCode
	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return accrualResp, err
		}

		var accrualOrder models.AccrualOrder
		if err := json.Unmarshal(body, &accrualOrder); err != nil {
			return accrualResp, err
		}

		accrualResp.AccrualOrder = &accrualOrder
		return accrualResp, nil

	case http.StatusNoContent:
		return accrualResp, ErrAccrualOrderNotRegistered

	case http.StatusTooManyRequests:
		accrualResp.RetryInterval = defaultRetryInterval
		if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
			if timeSeconds, err := strconv.Atoi(retryHeader); err == nil {
				accrualResp.RetryInterval = time.Duration(timeSeconds) * time.Second
			}
		}
		return accrualResp, ErrAccrualTooManyRequests

	case http.StatusInternalServerError:
		return accrualResp, ErrAccrualInternalError

	default:
		return accrualResp, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}
