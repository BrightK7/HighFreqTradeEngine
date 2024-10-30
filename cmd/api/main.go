package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cobo.leon.net/internal/data"
	"github.com/redis/go-redis/v9"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	config *config
	models data.Models
	logger *log.Logger
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	client := initRedisClient()
	defer client.Close()

	app := &application{
		config: &cfg,
		models: data.NewModels(client),
		logger: log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
	}

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("Starting server on %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

func initRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		DB:       0,
		Password: "",
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return client

}
