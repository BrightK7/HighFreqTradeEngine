package main

import (
	"flag"

	"github.com/redis/go-redis/v9"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	config *config
	models *models.Models
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	client := initRedisClient()
	defer client.Close()

	app := newApp(&cfg, client)
	err := app.run()
}

func initRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		DB:       0,
		Password: "",
	})

}
