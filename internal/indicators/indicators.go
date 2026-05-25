package indicators

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Sentinel errors returned by all indicator functions.
var (
	ErrInsufficientData = errors.New("insufficient data")
	ErrInvalidPeriod    = errors.New("period must be positive")
	ErrZeroDenominator  = errors.New("zero denominator")
)

// SMA returns the simple moving average of the last period prices.
func SMA(prices []float64, period int) (float64, error) {
	if period <= 0 {
		return 0, ErrInvalidPeriod
	}
	if len(prices) < period {
		return 0, fmt.Errorf("%w: SMA(%d) needs %d prices, have %d", ErrInsufficientData, period, period, len(prices))
	}
	var sum float64
	for _, p := range prices[len(prices)-period:] {
		sum += p
	}
	return sum / float64(period), nil
}

// EMA returns the exponential moving average of the last value in the series.
func EMA(prices []float64, period int) (float64, error) {
	if period <= 0 {
		return 0, ErrInvalidPeriod
	}
	if len(prices) < period {
		return 0, fmt.Errorf("%w: EMA(%d) needs %d prices, have %d", ErrInsufficientData, period, period, len(prices))
	}
	series := emaFull(prices, period)
	return series[len(series)-1], nil
}

// RSI returns the Relative Strength Index using Wilder's smoothing method.
func RSI(prices []float64, period int) (float64, error) {
	if period <= 0 {
		return 0, ErrInvalidPeriod
	}
	if len(prices) < period+1 {
		return 0, fmt.Errorf("%w: RSI(%d) needs %d prices, have %d", ErrInsufficientData, period, period+1, len(prices))
	}

	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		diff := prices[i] - prices[i-1]
		if diff > 0 {
			avgGain += diff
		} else {
			avgLoss -= diff
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Wilder's smoothing over remaining prices.
	for i := period + 1; i < len(prices); i++ {
		diff := prices[i] - prices[i-1]
		gain := math.Max(diff, 0)
		loss := math.Max(-diff, 0)
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}

	if avgLoss == 0 {
		return 100, nil
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs)), nil
}

// BollingerResult contains Bollinger Band values and the price's position within the band.
type BollingerResult struct {
	Upper    float64
	Middle   float64
	Lower    float64
	// Position is where the last price sits: 0.0 = at lower band, 1.0 = at upper band.
	Position float64
}

// BollingerBands computes Bollinger Bands using the given period and std-dev multiplier.
func BollingerBands(prices []float64, period int, mult float64) (BollingerResult, error) {
	if period <= 0 {
		return BollingerResult{}, ErrInvalidPeriod
	}
	if len(prices) < period {
		return BollingerResult{}, fmt.Errorf("%w: Bollinger(%d) needs %d prices, have %d", ErrInsufficientData, period, period, len(prices))
	}

	mean, _ := SMA(prices, period)
	window := prices[len(prices)-period:]

	var variance float64
	for _, p := range window {
		d := p - mean
		variance += d * d
	}
	stddev := math.Sqrt(variance / float64(period))

	upper := mean + mult*stddev
	lower := mean - mult*stddev

	var pos float64
	if band := upper - lower; band > 0 {
		pos = (prices[len(prices)-1] - lower) / band
	}

	return BollingerResult{Upper: upper, Middle: mean, Lower: lower, Position: pos}, nil
}

// MACDResult holds MACD line, signal line, and histogram.
type MACDResult struct {
	Line, Signal, Histogram float64
}

// MACD computes the MACD indicator using standard 12/26/9 parameters.
func MACD(prices []float64) (MACDResult, error) {
	return MACDCustom(prices, 12, 26, 9)
}

// MACDCustom computes MACD with custom fast/slow/signal periods.
// The original implementation had an EMA alignment bug: EMA() returned a trimmed slice,
// causing the MACD series reconstruction to silently mis-index. This version aligns
// all EMAs to the full price index before computing the MACD line.
func MACDCustom(prices []float64, fast, slow, signalPeriod int) (MACDResult, error) {
	minLen := slow + signalPeriod
	if len(prices) < minLen {
		return MACDResult{}, fmt.Errorf("%w: MACD(%d,%d,%d) needs %d prices, have %d",
			ErrInsufficientData, fast, slow, signalPeriod, minLen, len(prices))
	}

	ema12 := emaFull(prices, fast)
	ema26 := emaFull(prices, slow)

	// MACD line is valid from index slow-1 (first index where ema26 is seeded).
	macdLen := len(prices) - (slow - 1)
	macdSeries := make([]float64, macdLen)
	for i := slow - 1; i < len(prices); i++ {
		macdSeries[i-(slow-1)] = ema12[i] - ema26[i]
	}

	if len(macdSeries) < signalPeriod {
		return MACDResult{}, fmt.Errorf("%w: need %d MACD values for signal, have %d",
			ErrInsufficientData, signalPeriod, len(macdSeries))
	}

	signalSeries := emaFull(macdSeries, signalPeriod)
	line := macdSeries[len(macdSeries)-1]
	sig := signalSeries[len(signalSeries)-1]

	return MACDResult{Line: line, Signal: sig, Histogram: line - sig}, nil
}

// ATR computes Average True Range using a close-to-close approximation.
// Without OHLCV data, true range is approximated as |close[i] − close[i−1]|.
func ATR(prices []float64, period int) (float64, error) {
	if period <= 0 {
		return 0, ErrInvalidPeriod
	}
	if len(prices) < period+1 {
		return 0, fmt.Errorf("%w: ATR(%d) needs %d prices, have %d", ErrInsufficientData, period, period+1, len(prices))
	}

	tr := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		tr[i-1] = math.Abs(prices[i] - prices[i-1])
	}

	var seed float64
	for i := 0; i < period; i++ {
		seed += tr[i]
	}
	atr := seed / float64(period)

	for i := period; i < len(tr); i++ {
		atr = (atr*float64(period-1) + tr[i]) / float64(period)
	}
	return atr, nil
}

// StochasticResult holds %K and %D stochastic oscillator values.
type StochasticResult struct {
	K, D float64
}

// Stochastic computes the Stochastic Oscillator (%K fast, %D slow).
// Without OHLCV data, the highest and lowest closing prices within each window
// are used in place of the session high/low.
func Stochastic(prices []float64, kPeriod, dPeriod int) (StochasticResult, error) {
	if kPeriod <= 0 || dPeriod <= 0 {
		return StochasticResult{}, ErrInvalidPeriod
	}
	need := kPeriod + dPeriod - 1
	if len(prices) < need {
		return StochasticResult{}, fmt.Errorf("%w: Stochastic(%d,%d) needs %d prices, have %d",
			ErrInsufficientData, kPeriod, dPeriod, need, len(prices))
	}

	kValues := make([]float64, dPeriod)
	for j := 0; j < dPeriod; j++ {
		start := len(prices) - need + j
		window := prices[start : start+kPeriod]

		lo, hi := window[0], window[0]
		for _, p := range window[1:] {
			if p < lo {
				lo = p
			}
			if p > hi {
				hi = p
			}
		}

		if hi == lo {
			kValues[j] = 50
		} else {
			kValues[j] = (window[len(window)-1] - lo) / (hi - lo) * 100
		}
	}

	k := kValues[len(kValues)-1]
	var dSum float64
	for _, kv := range kValues {
		dSum += kv
	}
	return StochasticResult{K: k, D: dSum / float64(dPeriod)}, nil
}

// RegressionResult holds least-squares regression parameters and goodness of fit.
type RegressionResult struct {
	Slope, Intercept float64
	// R2 is the coefficient of determination: 1.0 = perfect fit, 0.0 = no explanatory power.
	R2 float64
	// Forecast is the predicted value at x = n (one step beyond the last observation).
	Forecast float64
}

// LinearRegression fits a least-squares line to prices and forecasts one step ahead.
func LinearRegression(prices []float64) (RegressionResult, error) {
	n := len(prices)
	if n < 3 {
		return RegressionResult{}, fmt.Errorf("%w: regression needs at least 3 prices, have %d", ErrInsufficientData, n)
	}

	var sumX, sumY, sumXY, sumX2 float64
	fn := float64(n)
	for i, p := range prices {
		x := float64(i)
		sumX += x
		sumY += p
		sumXY += x * p
		sumX2 += x * x
	}

	denom := fn*sumX2 - sumX*sumX
	if denom == 0 {
		return RegressionResult{}, ErrZeroDenominator
	}

	slope := (fn*sumXY - sumX*sumY) / denom
	intercept := (sumY - slope*sumX) / fn
	forecast := slope*fn + intercept

	meanY := sumY / fn
	var ssTot, ssRes float64
	for i, p := range prices {
		pred := slope*float64(i) + intercept
		ssRes += (p - pred) * (p - pred)
		ssTot += (p - meanY) * (p - meanY)
	}

	r2 := 0.0
	if ssTot > 0 {
		r2 = math.Max(0, 1-ssRes/ssTot)
	}

	return RegressionResult{Slope: slope, Intercept: intercept, R2: r2, Forecast: forecast}, nil
}

// emaFull returns a full-length EMA series aligned to prices.
// Indices [0, period-2] are zero (insufficient history); valid values start at period-1.
func emaFull(prices []float64, period int) []float64 {
	result := make([]float64, len(prices))
	k := 2.0 / float64(period+1)

	var seed float64
	for i := 0; i < period; i++ {
		seed += prices[i]
	}
	result[period-1] = seed / float64(period)

	for i := period; i < len(prices); i++ {
		result[i] = prices[i]*k + result[i-1]*(1-k)
	}
	return result
}

// FmtMoney formats a float as USD with thousand separators (e.g. $67,420.00).
func FmtMoney(v float64) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("$%.2fM", v/1_000_000)
	}
	s := fmt.Sprintf("%.2f", v)
	dot := strings.Index(s, ".")
	intStr, fracStr := s[:dot], s[dot:]
	var out strings.Builder
	out.WriteByte('$')
	for i, c := range intStr {
		if i > 0 && (len(intStr)-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteRune(c)
	}
	out.WriteString(fracStr)
	return out.String()
}
