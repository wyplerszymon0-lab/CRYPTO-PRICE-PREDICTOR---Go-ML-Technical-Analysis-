package report

import (
	"fmt"
	"strings"

	"github.com/wyplerszymon0-lab/crypto-predictor/internal/predictor"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	white  = "\033[97m"
)

func Print(symbol string, r predictor.Result) {
	signalColor := signalColor(r.Signal)

	fmt.Printf("%s%s═══ %s ═══%s\n", bold, cyan, strings.ToUpper(symbol), reset)
	fmt.Printf("  %-20s %s$%.2f%s\n", "Current Price:", white, r.CurrentPrice, reset)
	fmt.Printf("  %-20s %s$%.2f%s\n", "Predicted (next day):", white, r.PredictedPrice, reset)

	changeColor := green
	changeSign := "+"
	if r.ChangePercent < 0 {
		changeColor = red
		changeSign = ""
	}
	fmt.Printf("  %-20s %s%s%.2f%%%s\n", "Expected Change:", changeColor, changeSign, r.ChangePercent, reset)

	fmt.Println()
	fmt.Printf("  %-20s %s%s%s%s\n", "Signal:", bold, signalColor, r.Signal, reset)
	fmt.Printf("  %-20s %.1f%%\n", "Confidence:", r.Confidence)

	fmt.Println()
	fmt.Printf("  %s── Indicators ──%s\n", cyan, reset)
	fmt.Printf("  %-20s %.2f  %s\n", "RSI (14):", r.RSI, rsiLabel(r.RSI))
	fmt.Printf("  %-20s %.4f\n", "MACD Line:", r.MACD)
	fmt.Printf("  %-20s %.1f%%\n", "Bollinger Position:", r.BollingerPos)

	trendStr := green + "↑ BULLISH" + reset
	if r.Trend < 0 {
		trendStr = red + "↓ BEARISH" + reset
	}
	fmt.Printf("  %-20s %s\n", "SMA Trend:", trendStr)
	fmt.Printf("  %-20s %.3f\n", "Composite Score:", r.Score)
}

func signalColor(s predictor.Signal) string {
	switch s {
	case predictor.StrongBuy, predictor.Buy:
		return green
	case predictor.StrongSell, predictor.Sell:
		return red
	default:
		return yellow
	}
}

func rsiLabel(rsi float64) string {
	switch {
	case rsi < 30:
		return green + "(Oversold)" + reset
	case rsi > 70:
		return red + "(Overbought)" + reset
	default:
		return yellow + "(Neutral)" + reset
	}
}
```

---

**`go.mod`**
```
module github.com/wyplerszymon0-lab/crypto-predictor

go 1.22
