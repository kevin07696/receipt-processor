package main

import (
	"crypto/sha256"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	"github.com/kevin07696/receipt-processor/adapters/caches"
	receiptDomain "github.com/kevin07696/receipt-processor/domain/receipt"
	"github.com/kevin07696/receipt-processor/handlers"
	"github.com/kevin07696/receipt-processor/handlers/admin"
	receiptHandlers "github.com/kevin07696/receipt-processor/handlers/receipt"
	"github.com/kevin07696/receipt-processor/infrastructure/config"
	"github.com/kevin07696/receipt-processor/infrastructure/loggers"
)

func main() {
	env := config.LoadEnvConfig()

	h := loggers.ContextHandler{}
	if env.AppEnv == "prod" {
		h.Handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo})
	} else {
		h.Handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(h)
	slog.SetDefault(logger)

	cache := caches.NewLRUCache(env.CacheCap)
	var repository receiptDomain.IReceiptProcessorRepository = receiptDomain.NewReceiptProcessorRepository(&cache)

	env.Options.GenerateID = func(input string) string {
		if len(input) == 0 {
			return uuid.NewString()
		}

		hash := sha256.Sum256([]byte(input))
		hashBytes := hash[:16]
		hashUUID, err := uuid.FromBytes(hashBytes)
		if err != nil {
			return uuid.NewString()
		}

		return hashUUID.String()
	}

	receiptAPI := receiptDomain.NewReceiptProcessorService(repository, env.Options, env.Multipliers)

	receiptRouter := http.NewServeMux()
	receiptHandlers.InitializeRoutes(receiptRouter, &receiptAPI)
	receiptHandler := handlers.ChainMiddlewaresToHandler(receiptRouter, handlers.RequestIDMiddleware, handlers.RequestLoggerMiddleware)

	// TODO: router still is being accessed by host port
	adminRouter := http.NewServeMux()
	admin.InitializeRoutes(adminRouter)
	adminHandler := handlers.ChainMiddlewaresToHandler(adminRouter, handlers.RequestIDMiddleware, handlers.RequestLoggerMiddleware)

	go handlers.StartServer(env.AppPort, receiptHandler)
	go handlers.StartServer(env.AdminPort, adminHandler)

	// Wait for an interrupt signal to gracefully shutdown the application
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT)
	<-quit
	log.Println("Shutting down servers...")
}
