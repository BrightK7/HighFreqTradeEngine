package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cobo.leon.net/internal/data"
)

func (app *application) createOrder(w http.ResponseWriter, r *http.Request) {
	var orderReq data.OrderRequest
	err := json.NewDecoder(r.Body).Decode(&orderReq)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	if err = data.ValidateOrderRequest(&orderReq); err != nil {
		app.errorJSON(w, err)
		return
	}

	order := &data.Order{
		ID:        orderReq.ID,
		Price:     orderReq.Price,
		Quantity:  orderReq.Quantity,
		Timestamp: time.Now().Unix(),
	}

	if orderReq.Type == data.LimitOrder {
		err = app.models.Orders.AddLimitOrder(order)
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	} else if orderReq.Type == data.MarketOrder {
		if orderReq.Side == data.Buy {
			err = app.models.Orders.MatchBuyMarketOrder(*order)
			if err != nil {
				app.errorJSON(w, err)
				return
			}
		} else if orderReq.Side == data.Sell {
			err = app.models.Orders.MatchSellMarketOrder(*order)
			if err != nil {
				app.errorJSON(w, err)
				return
			}
		}
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Order received: %s", order.ID)
}
