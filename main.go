package main

import (
	"crypto/sha256"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/kevin07696/receipt-processor/adapters/caches"
	"github.com/kevin07696/receipt-processor/adapters/loggers"
	dReceipt "github.com/kevin07696/receipt-processor/domain/receipt"
	"github.com/kevin07696/receipt-processor/handlers"
	hReceipt "github.com/kevin07696/receipt-processor/handlers/receipt"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	env := loadEnvConfig()

	h := loggers.ContextHandler{}
	if env.AppEnv == "prod" {
		h.Handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo})
	} else {
		h.Handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(h)
	slog.SetDefault(logger)

	cache := caches.NewLRUCache(env.CacheCap)
	var repository dReceipt.IReceiptProcessorRepository = dReceipt.NewReceiptProcessorRepository(&cache)

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

	receiptAPI := dReceipt.NewReceiptProcessorService(repository, env.Options, env.Multipliers)

	router := http.NewServeMux()

	hReceipt.Handle(router, &receiptAPI)

	app := handlers.NewApp(env.Port, router)

	log.Printf("Starting server at :%d\n", env.Port)

	err = app.Run(handlers.RequestLoggerMiddleware, handlers.RequestIDMiddleware)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type Config struct {
	AppEnv      string
	Port        int
	CacheCap    int
	Multipliers dReceipt.Multipliers
	Options     dReceipt.Options
}

func loadEnvConfig() Config {
	env := map[string]interface{}{
		"APP_ENV":              "",
		"APP_PORT":             int(0),
		"MULT_RECEIPT":         int64(0),
		"MULT_ROUND_TOTAL":     int64(0),
		"MULT_DIVISIBLE_TOTAL": int64(0),
		"MULT_ITEMS":           float64(0),
		"MULT_DESCRIPTION":     float64(0),
		"MULT_PURCHASE_TIME":   int64(0),
		"MULT_PURCHASE_DATE":   int64(0),
		"START_TIME":           "",
		"END_TIME":             "",
		"TOTAL_MULTIPLE":       float64(0),
		"ITEMS_MULTIPLE":       int64(0),
		"DESCRIPTION_MULTIPLE": int64(0),
		"CACHE_CAP":            int(0),
	}

	for k := range env {
		val := os.Getenv(k)
		if val == "" {
			log.Printf("Environment variable %s is missing", k)
			continue
		}

		switch env[k].(type) {
		case string:
			env[k] = val
		case int:
			if parsedVal, err := strconv.ParseInt(val, 10, 64); err == nil {
				env[k] = int(parsedVal)
			} else {
				log.Fatalf("Error parsing %s: %v", k, err)
			}
		case int64:
			if parsedVal, err := strconv.ParseUint(val, 10, 16); err == nil {
				env[k] = int64(parsedVal)
			} else {
				log.Fatalf("Error parsing %s: %v", k, err)
			}
		case float64:
			if parsedVal, err := strconv.ParseFloat(val, 64); err == nil {
				env[k] = parsedVal
			} else {
				log.Fatalf("Error parsing %s: %v", k, err)
			}
		default:
			log.Fatalf("Unsupported type for environment variable %s", k)
		}
	}

	config := Config{
		AppEnv:   env["APP_ENV"].(string),
		Port:     env["APP_PORT"].(int),
		CacheCap: env["CACHE_CAP"].(int),
		Multipliers: dReceipt.Multipliers{
			Retailer:       env["MULT_RECEIPT"].(int64),
			RoundTotal:     env["MULT_ROUND_TOTAL"].(int64),
			DivisibleTotal: env["MULT_DIVISIBLE_TOTAL"].(int64),
			Items:          env["MULT_ITEMS"].(float64),
			Description:    env["MULT_DESCRIPTION"].(float64),
			PurchaseTime:   env["MULT_PURCHASE_TIME"].(int64),
			PurchaseDate:   env["MULT_PURCHASE_DATE"].(int64),
		},
		Options: dReceipt.Options{
			StartPurchaseTime:   env["START_TIME"].(string),
			EndPurchaseTime:     env["END_TIME"].(string),
			TotalMultiple:       env["TOTAL_MULTIPLE"].(float64),
			ItemsMultiple:       env["ITEMS_MULTIPLE"].(int64),
			DescriptionMultiple: env["DESCRIPTION_MULTIPLE"].(int64),
		},
	}

	return config
}
