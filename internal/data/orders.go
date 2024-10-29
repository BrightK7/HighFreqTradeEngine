package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderSide string
type OrderType string

const (
	Buy  OrderSide = "buy"
	Sell OrderSide = "sell"
)

const (
	LimitOrder  OrderType = "limit"
	MarketOrder OrderType = "market"
)

type Order struct {
	ID        string    `json:"id"`
	Side      OrderSide `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Timestamp int64     `json:"timestamp"`
}

type OrderModel struct {
	DB *redis.Client
}

func (m OrderModel) AddLimitOrder(order *Order) error {
	script := getAddOrderLuaScript()
	// save order to redis
	orderKey := fmt.Sprintf("order:%s", order.ID)
	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// save order into order book
	var orderBookKey string
	var score float64
	if order.Side == Buy {
		orderBookKey = "order_book:buy"
		score = -order.Price
	} else if order.Side == Sell {
		orderBookKey = "order_book:sell"
		score = order.Price
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = m.DB.Eval(ctx, script, []string{orderKey, orderBookKey}, string(orderData), score, order.ID).Result()
	if err != nil {
		return fmt.Errorf("error executing Lua script: %w", err)
	}
	return nil
}

func (m OrderModel) GetOrder(id string) (Order, error) {
	return Order{}, nil
}

func (m OrderModel) UpdateOrder(order Order) error {
	return nil
}

func (m OrderModel) Delete(id string) error {
	return nil
}

func (m OrderModel) matchMarketOrder(order Order) error {
	script := getMatchOrderLuaScript()

	var orderBookKey string
	if order.Side == Buy {
		// buy order matches with sell order
		orderBookKey = "order_book:sell"
	} else if order.Side == Sell {
		// buy order matches with sell order
		orderBookKey = "order_book:buy"
	}

	orderData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	_, err = m.DB.Eval(context.Background(), script, []string{orderBookKey}, string(orderData), order.Price, order.ID).Result()
	if err != nil {
		return fmt.Errorf("error executing Lua script: %w", err)
	}
	return nil
}

func (m OrderModel) MatchBuyMarketOrder(order Order) error {
	return nil
}

func (m OrderModel) MatchSellMarketOrder(order Order) error {
	return nil
}

func getAddOrderLuaScript() string {
	return `
		local orderKey = KEYS[1]
		local orderBookKey = KEYS[2]
		local orderData = ARGV[1]
		local score = ARGV[2]
		local orderId = ARGV[3]

		-- check if order already exists
		if redis.call("EXISTS", orderKey) == 1 then
			return redis.error_reply("order already exists: " .. orderId)
		end

		-- save order
		local saved = redis.call("HSET", orderKey, "data", orderData)
		if saved == 0 then
			return redis.error_reply("failed to save order: " .. orderId)
		end

		-- save order in order book
		local added = redis.call("ZADD", orderBookKey, score, orderId)
		if added == 0 then
			return redis.error_reply("failed to add order to order book: " .. orderId)
		end

		return "Order added: " .. orderId
	`
}

func getMatchOrderLuaScript() string {
	return `
		local orderBookKey = KEYS[1]
		local orderData = ARGV[1]

		-- retrieve order data
		local order = cjson.decode(orderData)
		local orderQuantity = order["quantity"]
		local orderSide = order["side"]

		local function getBestOrder(orderBookKey, orderSide)
			if orderSide == "BUY" then
				return redis.call("ZRANGE", orderBookKey, 0, 0)[1]
			elseif orderSide == "SELL" then
				return redis.call("ZREVRANGE", orderBookKey, 0, 0)[1]
			end
		end

		local function processTrade(order, bestOrder, tradeQuantity)
			order["quantity"] = order["quantity"] - tradeQuantity
			bestOrder["quantity"] = bestOrder["quantity"] - tradeQuantity

			local trade = {
				buy_order_id = orderSide == "BUY" and order["id"] or bestOrder["id"],
				sell_order_id = orderSide == "SELL" and order["id"] or bestOrder["id"],
				price = bestOrder["price"],
				quantity = tradeQuantity,
				timestamp = redis.call("TIME")[1]
			}
			local tradeKey = "trade:" .. trade["timestamp"]
			redis.call("HSET", tradeKey, "data", cjson.encode(trade))
		end

		local function updateBestOrder(orderBookKey, bestOrderId, bestOrder)
			if bestOrder["quantity"] <= 0 then
				redis.call("ZREM", orderBookKey, bestOrderId)
				redis.call("DEL", "order:" .. bestOrderId)
			else
				local updateBestOrder = cjson.encode(bestOrder)
				redis.call("HSET", "order:" .. bestOrderId, "data", updateBestOrder)
			end
		end

		while orderQuantity > 0 do
			local bestOrderId = getBestOrderID(orderBookKey, orderSide)

			if not bestOrderId then
				break
			end

			local bestOrderKey = "order:" .. bestOrderId
			local bestOrderData = redis.call("HGET", bestOrderKey, "data")
			if not bestOrderData then
				redis.call("ZREM", orderBookKey, bestOrderId)
				break
			end

			local bestOrder = cjson.decode(bestOrderData)
			local tradeQuantity = math.min(orderQuantity, bestOrder["quantity"])

			processTrade(order, bestOrder, tradeQuantity)

			updateBestOrder(orderBookKey, bestOrderId, bestOrder)

			orderQuantity = orderQuantity - tradeQuantity
			if orderQuantity <= 0 then
				break
			end
		end
	`
}