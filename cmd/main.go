package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gqvz/mvc/pkg/api"
	"github.com/gqvz/mvc/pkg/config"
	"github.com/gqvz/mvc/pkg/models"

	_ "github.com/gqvz/mvc/docs"
)

// @title           MVC
// @version         1.0
// @description     MVC Assignment

// @host 	     mvc.gqvz.xyz
// @BasePath     /api
// @schemes      https

// @securityDefinitions.apikey jwt
// @in header
// @name Authorization
// @description JWT token in Authorization header
func main() {
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatal("error loading appConfig: ", err)
		return
	}

	_, err = models.InitDatabase(appConfig.DB)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
		return
	}

	router := api.CreateRouter(appConfig)

	server := &http.Server{
		Addr:    appConfig.ServerAddress,
		Handler: router,
	}

	go func() {
		fmt.Println("Starting server on", appConfig.ServerAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start http server: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := models.CloseDatabase(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited gracefully")
}
