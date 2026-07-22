package main

import (
	"context"
	"log"
	"time"

	"sistema-editorial/editora/backend/src/config"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := config.NewDatabasePool(ctx, cfg)
	if err != nil {
		log.Fatalf("erro ao inicializar backend: %v", err)
	}
	defer pool.Close()

	app := config.NewHTTPServer(cfg, pool)

	log.Printf("%s disponivel na porta %s", cfg.AppName, cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
