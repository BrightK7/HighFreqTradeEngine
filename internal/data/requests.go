package data

import "fmt"

func ValidateOrderRequest(orderReq *OrderRequest) error {
	if orderReq.ID == "" {
		return fmt.Errorf("order ID is required")
	}
	if orderReq.Side != Buy && orderReq.Side != Sell {
		return fmt.Errorf("invalid order side for %s", orderReq.Side)
	}
	if orderReq.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if orderReq.Type != "LIMIT" && orderReq.Type != "MARKET" {
		return fmt.Errorf("invalid order type: must be 'LIMIT' or 'MARKET', current is %s", orderReq.Type)
	}
	if orderReq.Type == "LIMIT" && orderReq.Price <= 0 {
		return fmt.Errorf("price must be positive for limit orders")
	}
	return nil
}
