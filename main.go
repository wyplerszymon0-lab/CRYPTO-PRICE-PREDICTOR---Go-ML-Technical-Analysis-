package main

import (
	"fmt"
	"os"

	"github.com/wyplerszymon0-lab/crypto-predictor/internal/api"
	"github.com/wyplerszymon0-lab/crypto-predictor/internal/predictor"
	"github.com/wyplerszymon0-lab/crypto-predictor/internal/report"
)

func main() {
	symbols := []string{"bitcoin", "ethereum", "solana"}

	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║     CRYPTO PRICE PREDICTOR v1.0      ║")
	fmt.Println("║     Go · ML · Technical Analysis     ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()

	client := api.NewCoinGeckoClient()

	for _, symbol := range symbols {
		fmt.Printf("▶ Fetching data for %s...\n", symbol)

		prices, err := client.FetchPrices(symbol, 30)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error fetching %s: %v\n", symbol, err)
			continue
		}

		engine := predictor.NewEngine(prices)
		result := engine.Predict()

		report.Print(symbol, result)
		fmt.Println()
	}
}
