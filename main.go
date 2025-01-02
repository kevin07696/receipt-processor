package main

import (
	"crypto/sha256"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/kevin07696/receipt-processor/adapters"
	dreceipt "github.com/kevin07696/receipt-processor/domain/receipt"
	"github.com/kevin07696/receipt-processor/handlers"
	hreceipt "github.com/kevin07696/receipt-processor/handlers/receipt"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	env := loadEnvConfig()

	cache := adapters.NewLRUCache(200000)
	var repository dreceipt.IReceiptProcessorRepository = dreceipt.NewReceiptProcessorRepository(&cache)

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

	receiptAPI := dreceipt.NewReceiptProcessorService(repository, env.Options, env.Multipliers)

	router := http.NewServeMux()

	hreceipt.Handle(router, &receiptAPI)

	app := handlers.NewApp(env.Port, router)

	log.Printf("Starting server at :%d\n", env.Port)
	if err := app.Run(handlers.RequestLoggerMiddleware); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type Config struct {
	AppEnv      string
	Port        int
	Multipliers dreceipt.Multipliers
	Options     dreceipt.Options
}

func loadEnvConfig() Config {
	env := map[string]interface{}{
		"APP_ENV":              "",
		"APP_PORT":             int(0),
		"MULT_RECEIPT":         uint16(0),
		"MULT_ROUND_TOTAL":     uint16(0),
		"MULT_DIVISIBLE_TOTAL": uint16(0),
		"MULT_ITEMS":           float64(0),
		"MULT_DESCRIPTION":     float64(0),
		"MULT_PURCHASE_TIME":   uint16(0),
		"MULT_PURCHASE_DATE":   uint16(0),
		"START_TIME":           "",
		"END_TIME":             "",
		"TOTAL_MULTIPLE":       float64(0),
		"ITEMS_MULTIPLE":       uint16(0),
		"DESCRIPTION_MULTIPLE": uint16(0),
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
		case uint16:
			if parsedVal, err := strconv.ParseUint(val, 10, 16); err == nil {
				env[k] = uint16(parsedVal)
			} else {
				log.Fatalf("Error parsing %s: %v", k, err)
			}
		case int:
			if parsedVal, err := strconv.ParseInt(val, 10, 64); err == nil {
				env[k] = int(parsedVal)
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
		AppEnv: env["APP_ENV"].(string),
		Port:   env["APP_PORT"].(int),
		Multipliers: dreceipt.Multipliers{
			Receipt:        env["MULT_RECEIPT"].(uint16),
			RoundTotal:     env["MULT_ROUND_TOTAL"].(uint16),
			DivisibleTotal: env["MULT_DIVISIBLE_TOTAL"].(uint16),
			Items:          env["MULT_ITEMS"].(float64),
			Description:    env["MULT_DESCRIPTION"].(float64),
			PurchaseTime:   env["MULT_PURCHASE_TIME"].(uint16),
			PurchaseDate:   env["MULT_PURCHASE_DATE"].(uint16),
		},
		Options: dreceipt.Options{
			StartPurchaseTime:   env["START_TIME"].(string),
			EndPurchaseTime:     env["END_TIME"].(string),
			TotalMultiple:       env["TOTAL_MULTIPLE"].(float64),
			ItemsMultiple:       env["ITEMS_MULTIPLE"].(uint16),
			DescriptionMultiple: env["DESCRIPTION_MULTIPLE"].(uint16),
		},
	}

	return config
}
