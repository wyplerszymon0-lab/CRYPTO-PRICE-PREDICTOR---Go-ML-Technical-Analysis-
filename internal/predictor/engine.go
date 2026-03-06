package predictor

import (
	"math"

	"github.com/wyplerszymon0-lab/crypto-predictor/internal/indicators"
)

type Signal string

const (
	StrongBuy  Signal = "STRONG BUY"
	Buy        Signal = "BUY"
	Neutral    Signal = "NEUTRAL"
	Sell       Signal = "SELL"
	StrongSell Signal = "STRONG SELL"
)

type Result struct {
	CurrentPrice   float64
	PredictedPrice float64
	ChangePercent  float64
	Signal         Signal
	Confidence     float64
	RSI            float64
	MACD           float64
	BollingerPos   float64
	Trend          float64
	Score          float64
}

type Engine struct {
	prices []float64
}

func NewEngine(prices []float64) *Engine {
	return &Engine{prices: prices}
}

func (e *Engine) Predict() Result {
	prices := e.prices
	current := prices[len(prices)-1]

	rsi := indicators.RSI(prices, 14)
	macd := indicators.MACDIndicator(prices)
	bb := indicators.Bollinger(prices, 20, 2.0)
	sma7 := indicators.SMA(prices, 7)
	sma20 := indicators.SMA(prices, 20)

	score := 0.0
	weights := 0.0

	rsiScore := 0.0
	switch {
	case rsi < 30:
		rsiScore = 1.0
	case rsi < 45:
		rsiScore = 0.5
	case rsi > 70:
		rsiScore = -1.0
	case rsi > 55:
		rsiScore = -0.5
	}
	score += rsiScore * 0.25
	weights += 0.25

	macdScore := 0.0
	if macd.Line != 0 {
		if macd.Hist > 0 && macd.Line > 0 {
			macdScore = 1.0
		} else if macd.Hist > 0 {
			macdScore = 0.5
		} else if macd.Hist < 0 && macd.Line < 0 {
			macdScore = -1.0
		} else {
			macdScore = -0.5
		}
	}
	score += macdScore * 0.25
	weights += 0.25

	bbRange := bb.Upper - bb.Lower
	bbPos := 0.0
	if bbRange > 0 {
		bbPos = (current - bb.Lower) / bbRange
	}
	bbScore := -(bbPos*2 - 1)
	score += bbScore * 0.20
	weights += 0.20

	trendScore := 0.0
	if sma7 != nil && sma20 != nil {
		s7 := sma7[len(sma7)-1]
		s20 := sma20[len(sma20)-1]
		if s7 > s20 {
			trendScore = 1.0
		} else {
			trendScore = -1.0
		}
	}
	score += trendScore * 0.30
	weights += 0.30

	normalized := score / weights

	predicted := linearRegression(prices[len(prices)-7:])
	changePct := (predicted - current) / current * 100

	signal := scoreToSignal(normalized)
	confidence := math.Abs(normalized) * 100

	return Result{
		CurrentPrice:   current,
		PredictedPrice: predicted,
		ChangePercent:  changePct,
		Signal:         signal,
		Confidence:     confidence,
		RSI:            rsi,
		MACD:           macd.Line,
		BollingerPos:   bbPos * 100,
		Trend:          trendScore,
		Score:          normalized,
	}
}

func scoreToSignal(score float64) Signal {
	switch {
	case score >= 0.6:
		return StrongBuy
	case score >= 0.2:
		return Buy
	case score <= -0.6:
		return StrongSell
	case score <= -0.2:
		return Sell
	default:
		return Neutral
	}
}

func linearRegression(prices []float64) float64 {
	n := float64(len(prices))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, p := range prices {
		x := float64(i)
		sumX += x
		sumY += p
		sumXY += x * p
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	return slope*n + intercept
}
