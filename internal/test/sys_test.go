package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var ctx = context.Background()

// Redis client setup (modify as needed)
var redisClient *redis.Client

func init() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis server address
	})
}

// Test placing a limit order
func TestPlaceLimitOrder(t *testing.T) {
	order := map[string]interface{}{
		"id":       "1",
		"price":    100,
		"quantity": 10,
		"side":     "BUY",
	}

	orderJSON, _ := json.Marshal(order)
	req, err := http.NewRequest("POST", "/v1/orders", bytes.NewBuffer(orderJSON))
	if err != nil {
		t.Fatal(err)
	}

	

	// Verify the order is stored in Redis
	exists, err := redisClient.Exists(ctx, "order:1").Result()
	if err != nil {
		t.Fatal(err)
	}
	if exists != 1 {
		t.Errorf("order was not created in Redis")
	}
}

// Test placing a market order
func TestPlaceMarketOrder(t *testing.T) {
	// Place a limit order to match against
	order := map[string]interface{}{
		"id":       "2",
		"price":    100,
		"quantity": 10,
		"side":     "SELL",
	}
	orderJSON, _ := json.Marshal(order)
	_ = http.Post("/orders/limit", "application/json", bytes.NewBuffer(orderJSON))

	marketOrder := map[string]interface{}{
		"id":       "3",
		"quantity": 5,
		"side":     "BUY",
	}

	marketOrderJSON, _ := json.Marshal(marketOrder)
	req, err := http.NewRequest("POST", "/v1/orders", bytes.NewBuffer(marketOrderJSON))
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(PlaceMarketOrder) // Your market order handler
	handler.ServeHTTP(rr, req)

	// Check the response status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify that the trade was created
	tradeKey := "trade:" + "timestamp" // Adjust based on how you store trades
	exists, err := redisClient.Exists(ctx, tradeKey).Result()
	if err != nil {
		t.Fatal(err)
	}
	if exists != 1 {
		t.Errorf("trade was not created in Redis")
	}
}

// Cleanup Redis after tests
func cleanup() {
	redisClient.FlushDB(ctx)
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup Redis
	cleanup()

	// Exit with code
	os.Exit(code)
}
