package data

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Orders interface {
		AddLimitOrder(order *Order) error
		MatchBuyMarketOrder(order Order) error
		matchMarketOrder(order Order) error
		MatchSellMarketOrder(order Order) error
		GetOrder(id string) (Order, error)
		UpdateOrder(order Order) error
	}
}

func NewModels(db *redis.Client) Models {
	return Models{
		Orders: OrderModel{DB: db},
	}
}

type OrderRequest struct {
	ID       string    `json:"id"`
	Side     OrderSide `json:"side"`
	Type     OrderType `json:"type"`
	Price    float64   `json:"price,omitempty"`
	Quantity float64   `json:"quantity"`
}
