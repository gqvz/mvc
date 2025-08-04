package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gqvz/mvc/pkg/models"
	"github.com/joho/godotenv"
	"go-simpler.org/env"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var config struct {
	JwtSecret string `env:"JWT_SECRET"`
	DB        models.DBConfig
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading environment variables: ", err)
		return
	}
	if err := env.Load(&config, nil); err != nil {
		log.Fatal("failed to load config: ", err)
		return
	}
	_, err = models.InitDatabase(config.DB)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
		return
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})
	go func() {
		err = app.Listen(":3000")
		if err != nil {
			log.Fatal("Failed to start http server: %v", err)
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Create a deadline for server shutdown
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := app.Shutdown(); err != nil {
		log.Info("Server forced to shutdown: %v", err)
	}

	// Close database connection
	if err := models.CloseDatabase(); err != nil {
		log.Info("Error closing database: %v", err)
	}

	log.Info("Server exited gracefully")
}
