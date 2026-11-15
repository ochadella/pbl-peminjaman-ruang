package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/ochadella/pbl-peminjaman-ruang/config"
	"github.com/ochadella/pbl-peminjaman-ruang/database"
	"github.com/ochadella/pbl-peminjaman-ruang/route"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET belum diisi di .env")
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	app := fiber.New()

	route.SetupRoutes(app, db, cfg.JWTSecret)

	log.Fatal(app.Listen(":" + cfg.AppPort))
}
