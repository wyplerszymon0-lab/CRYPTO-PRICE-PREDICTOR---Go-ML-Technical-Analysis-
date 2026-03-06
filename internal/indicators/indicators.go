package indicators

import "math"

func SMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}
	result := make([]float64, len(prices)-period+1)
	for i := range result {
		sum := 0.0
		for _, p := range prices[i : i+period] {
			sum += p
		}
		result[i] = sum / float64(period)
	}
	return result
}

func EMA(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}
	k := 2.0 / float64(period+1)
	result := make([]float64, len(prices))

	sum := 0.0
	for _, p := range prices[:period] {
		sum += p
	}
	result[period-1] = sum / float64(period)

	for i := period; i < len(prices); i++ {
		result[i] = prices[i]*k + result[i-1]*(1-k)
	}
	return result[period-1:]
}

func RSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50
	}

	gains, losses := 0.0, 0.0
	for i := 1; i <= period; i++ {
		diff := prices[i] - prices[i-1]
		if diff > 0 {
			gains += diff
		} else {
			losses -= diff
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	for i := period + 1; i < len(prices); i++ {
		diff := prices[i] - prices[i-1]
		if diff > 0 {
			avgGain = (avgGain*float64(period-1) + diff) / float64(period)
			avgLoss = (avgLoss * float64(period-1)) / float64(period)
		} else {
			avgGain = (avgGain * float64(period-1)) / float64(period)
			avgLoss = (avgLoss*float64(period-1) - diff) / float64(period)
		}
	}

	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

type BollingerBands struct {
	Upper  float64
	Middle float64
	Lower  float64
}

func Bollinger(prices []float64, period int, multiplier float64) BollingerBands {
	sma := SMA(prices, period)
	if sma == nil {
		return BollingerBands{}
	}
	mid := sma[len(sma)-1]

	window := prices[len(prices)-period:]
	variance := 0.0
	for _, p := range window {
		diff := p - mid
		variance += diff * diff
	}
	stddev := math.Sqrt(variance / float64(period))

	return BollingerBands{
		Upper:  mid + multiplier*stddev,
		Middle: mid,
		Lower:  mid - multiplier*stddev,
	}
}

type MACD struct {
	Line   float64
	Signal float64
	Hist   float64
}

func MACDIndicator(prices []float64) MACD {
	if len(prices) < 26 {
		return MACD{}
	}
	ema12 := EMA(prices, 12)
	ema26 := EMA(prices, 26)
	if len(ema12) == 0 || len(ema26) == 0 {
		return MACD{}
	}

	macdLine := ema12[len(ema12)-1] - ema26[len(ema26)-1]

	macdSeries := make([]float64, len(ema26))
	for i := range ema26 {
		idx12 := len(ema12) - len(ema26) + i
		if idx12 >= 0 {
			macdSeries[i] = ema12[idx12] - ema26[i]
		}
	}

	signalEMA := EMA(macdSeries, 9)
	signal := 0.0
	if len(signalEMA) > 0 {
		signal = signalEMA[len(signalEMA)-1]
	}

	return MACD{
		Line:   macdLine,
		Signal: signal,
		Hist:   macdLine - signal,
	}
}
