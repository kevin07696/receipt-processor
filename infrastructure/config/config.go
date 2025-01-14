package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	receiptDomain "github.com/kevin07696/receipt-processor/domain/receipt"
)

type Config struct {
	AppEnv      string
	AppPort     int
	AdminPort   int
	CacheCap    int
	Multipliers receiptDomain.Multipliers
	Options     receiptDomain.Options
}

func LoadEnvConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	env := map[string]interface{}{
		"APP_ENV":              "",
		"APP_PORT":             int(0),
		"ADMIN_PORT":           int(0),
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
		AppEnv:    env["APP_ENV"].(string),
		AppPort:   env["APP_PORT"].(int),
		AdminPort: env["ADMIN_PORT"].(int),
		CacheCap:  env["CACHE_CAP"].(int),
		Multipliers: receiptDomain.Multipliers{
			Retailer:       env["MULT_RECEIPT"].(int64),
			RoundTotal:     env["MULT_ROUND_TOTAL"].(int64),
			DivisibleTotal: env["MULT_DIVISIBLE_TOTAL"].(int64),
			Items:          env["MULT_ITEMS"].(float64),
			Description:    env["MULT_DESCRIPTION"].(float64),
			PurchaseTime:   env["MULT_PURCHASE_TIME"].(int64),
			PurchaseDate:   env["MULT_PURCHASE_DATE"].(int64),
		},
		Options: receiptDomain.Options{
			StartPurchaseTime:   env["START_TIME"].(string),
			EndPurchaseTime:     env["END_TIME"].(string),
			TotalMultiple:       env["TOTAL_MULTIPLE"].(float64),
			ItemsMultiple:       env["ITEMS_MULTIPLE"].(int64),
			DescriptionMultiple: env["DESCRIPTION_MULTIPLE"].(int64),
		},
	}

	return config
}
