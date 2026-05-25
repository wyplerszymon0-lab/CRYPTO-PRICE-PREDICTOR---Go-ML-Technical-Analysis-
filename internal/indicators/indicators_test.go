package indicators

import (
	"errors"
	"math"
	"testing"
)

const eps = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < eps
}

// ── SMA ──────────────────────────────────────────────────────────────────────

func TestSMA(t *testing.T) {
	tests := []struct {
		name    string
		prices  []float64
		period  int
		want    float64
		wantErr error
	}{
		{
			name:   "last 3 of 5",
			prices: []float64{1, 2, 3, 4, 5},
			period: 3,
			want:   4.0, // (3+4+5)/3
		},
		{
			name:   "full series",
			prices: []float64{10, 20, 30},
			period: 3,
			want:   20.0,
		},
		{
			name:   "single element",
			prices: []float64{42},
			period: 1,
			want:   42.0,
		},
		{
			name:    "insufficient data",
			prices:  []float64{1, 2},
			period:  5,
			wantErr: ErrInsufficientData,
		},
		{
			name:    "invalid period zero",
			prices:  []float64{1, 2, 3},
			period:  0,
			wantErr: ErrInvalidPeriod,
		},
		{
			name:    "invalid period negative",
			prices:  []float64{1, 2, 3},
			period:  -1,
			wantErr: ErrInvalidPeriod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SMA(tt.prices, tt.period)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("SMA() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("SMA() unexpected error: %v", err)
			}
			if !almostEqual(got, tt.want) {
				t.Errorf("SMA() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ── EMA ──────────────────────────────────────────────────────────────────────

func TestEMA(t *testing.T) {
	t.Run("seeded correctly on exact period", func(t *testing.T) {
		// With only period prices, EMA == SMA.
		prices := []float64{1, 2, 3, 4, 5}
		got, err := EMA(prices, 5)
		if err != nil {
			t.Fatal(err)
		}
		sma, _ := SMA(prices, 5)
		if !almostEqual(got, sma) {
			t.Errorf("EMA with n=period: got %v, want SMA %v", got, sma)
		}
	})

	t.Run("converges toward latest price on flat series", func(t *testing.T) {
		// On a constant series, EMA == constant.
		prices := make([]float64, 50)
		for i := range prices {
			prices[i] = 100
		}
		got, _ := EMA(prices, 12)
		if math.Abs(got-100) > 1e-6 {
			t.Errorf("EMA on flat series: got %v, want 100", got)
		}
	})

	t.Run("insufficient data", func(t *testing.T) {
		_, err := EMA([]float64{1, 2}, 5)
		if !errors.Is(err, ErrInsufficientData) {
			t.Errorf("expected ErrInsufficientData, got %v", err)
		}
	})
}

// ── RSI ──────────────────────────────────────────────────────────────────────

func TestRSI(t *testing.T) {
	t.Run("all gains returns 100", func(t *testing.T) {
		// Strictly increasing prices → no losses → RSI = 100.
		prices := make([]float64, 20)
		for i := range prices {
			prices[i] = float64(i + 1)
		}
		got, err := RSI(prices, 14)
		if err != nil {
			t.Fatal(err)
		}
		if !almostEqual(got, 100) {
			t.Errorf("RSI(all-up) = %v, want 100", got)
		}
	})

	t.Run("all losses returns 0", func(t *testing.T) {
		// Strictly decreasing prices → no gains → RSI = 0.
		prices := make([]float64, 20)
		for i := range prices {
			prices[i] = float64(20 - i)
		}
		got, err := RSI(prices, 14)
		if err != nil {
			t.Fatal(err)
		}
		if !almostEqual(got, 0) {
			t.Errorf("RSI(all-down) = %v, want 0", got)
		}
	})

	t.Run("bounded in [0, 100]", func(t *testing.T) {
		prices := []float64{10, 20, 15, 30, 25, 40, 35, 50, 45, 60, 55, 70, 65, 80, 75}
		got, err := RSI(prices, 14)
		if err != nil {
			t.Fatal(err)
		}
		if got < 0 || got > 100 {
			t.Errorf("RSI out of [0,100]: %v", got)
		}
	})

	t.Run("insufficient data", func(t *testing.T) {
		_, err := RSI([]float64{1, 2, 3}, 14)
		if !errors.Is(err, ErrInsufficientData) {
			t.Errorf("expected ErrInsufficientData, got %v", err)
		}
	})
}

// ── Bollinger Bands ───────────────────────────────────────────────────────────

func TestBollingerBands(t *testing.T) {
	t.Run("position below 0.5 when price is below mean", func(t *testing.T) {
		// Last price is well below the 20-period mean → position < 0.5.
		prices := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100,
			100, 100, 100, 100, 100, 100, 100, 100, 100, 90}
		bb, err := BollingerBands(prices, 20, 2.0)
		if err != nil {
			t.Fatal(err)
		}
		if bb.Position >= 0.5 {
			t.Errorf("below-mean price should give position < 0.5, got %v", bb.Position)
		}
	})

	t.Run("flat series has zero std-dev", func(t *testing.T) {
		prices := make([]float64, 20)
		for i := range prices {
			prices[i] = 50
		}
		bb, err := BollingerBands(prices, 20, 2.0)
		if err != nil {
			t.Fatal(err)
		}
		if !almostEqual(bb.Upper, 50) || !almostEqual(bb.Lower, 50) {
			t.Errorf("flat series: upper/lower should both be 50, got %v/%v", bb.Upper, bb.Lower)
		}
	})

	t.Run("insufficient data", func(t *testing.T) {
		_, err := BollingerBands([]float64{1, 2, 3}, 20, 2.0)
		if !errors.Is(err, ErrInsufficientData) {
			t.Errorf("expected ErrInsufficientData, got %v", err)
		}
	})
}

// ── MACD ─────────────────────────────────────────────────────────────────────

func TestMACD(t *testing.T) {
	t.Run("histogram equals line minus signal", func(t *testing.T) {
		prices := make([]float64, 40)
		for i := range prices {
			prices[i] = 100 + float64(i)*0.5
		}
		got, err := MACD(prices)
		if err != nil {
			t.Fatal(err)
		}
		if !almostEqual(got.Histogram, got.Line-got.Signal) {
			t.Errorf("histogram (%v) != line (%v) - signal (%v)", got.Histogram, got.Line, got.Signal)
		}
	})

	t.Run("accelerating uptrend gives positive histogram", func(t *testing.T) {
		// An accelerating (quadratic) price series causes the MACD line to grow,
		// making the histogram (line − signal EMA) positive.
		prices := make([]float64, 40)
		for i := range prices {
			prices[i] = 100 + float64(i)*float64(i)*0.5 // quadratic
		}
		got, _ := MACD(prices)
		if got.Histogram <= 0 {
			t.Errorf("accelerating uptrend: expected positive histogram, got %v", got.Histogram)
		}
	})

	t.Run("insufficient data", func(t *testing.T) {
		_, err := MACD([]float64{1, 2, 3, 4, 5})
		if !errors.Is(err, ErrInsufficientData) {
			t.Errorf("expected ErrInsufficientData, got %v", err)
		}
	})
}

// ── ATR ───────────────────────────────────────────────────────────────────────

func TestATR(t *testing.T) {
	t.Run("positive on volatile series", func(t *testing.T) {
		prices := []float64{100, 110, 95, 115, 105, 120, 100, 130, 110, 125,
			105, 115, 95, 120, 100, 115}
		got, err := ATR(prices, 14)
		if err != nil {
			t.Fatal(err)
		}
		if got <= 0 {
			t.Errorf("ATR should be positive, got %v", got)
		}
	})

	t.Run("zero on flat series", func(t *testing.T) {
		prices := make([]float64, 20)
		for i := range prices {
			prices[i] = 50
		}
		got, _ := ATR(prices, 14)
		if got != 0 {
			t.Errorf("flat series ATR should be 0, got %v", got)
		}
	})

	t.Run("insufficient data", func(t *testing.T) {
		_, err := ATR([]float64{1, 2, 3}, 14)
		if !errors.Is(err, ErrInsufficientData) {
			t.Errorf("expected ErrInsufficientData, got %v", err)
		}
	})
}

// ── Stochastic ────────────────────────────────────────────────────────────────

func TestStochastic(t *testing.T) {
	t.Run("100 when last price equals period high", func(t *testing.T) {
		// All prices the same except the last which is highest.
		prices := make([]float64, 20)
		for i := range prices {
			prices[i] = 50
		}
		prices[len(prices)-1] = 100 // period high
		got, err := Stochastic(prices, 14, 3)
		if err != nil {
			t.Fatal(err)
		}
		if got.K != 100 {
			t.Errorf("stoch K should be 100, got %v", got.K)
		}
	})

	t.Run("0 when last price equals period low", func(t *testing.T) {
		prices := make([]float64, 20)
		for i := range prices {
			prices[i] = 50
		}
		prices[len(prices)-1] = 1 // period low
		got, err := Stochastic(prices, 14, 3)
		if err != nil {
			t.Fatal(err)
		}
		if got.K != 0 {
			t.Errorf("stoch K should be 0, got %v", got.K)
		}
	})

	t.Run("bounded in [0, 100]", func(t *testing.T) {
		prices := []float64{10, 20, 15, 30, 25, 40, 35, 50, 45, 60,
			55, 70, 65, 80, 75, 90, 85, 100}
		got, err := Stochastic(prices, 14, 3)
		if err != nil {
			t.Fatal(err)
		}
		if got.K < 0 || got.K > 100 || got.D < 0 || got.D > 100 {
			t.Errorf("K=%v D=%v out of [0,100]", got.K, got.D)
		}
	})
}

// ── Linear Regression ─────────────────────────────────────────────────────────

func TestLinearRegression(t *testing.T) {
	t.Run("perfect fit on y=2x+1", func(t *testing.T) {
		prices := make([]float64, 10)
		for i := range prices {
			prices[i] = 2*float64(i) + 1
		}
		got, err := LinearRegression(prices)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.R2-1.0) > 1e-9 {
			t.Errorf("R² = %v, want 1.0", got.R2)
		}
		if math.Abs(got.Slope-2.0) > 1e-9 {
			t.Errorf("slope = %v, want 2.0", got.Slope)
		}
		wantForecast := 2*float64(len(prices)) + 1
		if math.Abs(got.Forecast-wantForecast) > 1e-9 {
			t.Errorf("forecast = %v, want %v", got.Forecast, wantForecast)
		}
	})

	t.Run("flat series slope is zero", func(t *testing.T) {
		prices := make([]float64, 10)
		for i := range prices {
			prices[i] = 42
		}
		got, err := LinearRegression(prices)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.Slope) > 1e-9 {
			t.Errorf("flat series: slope = %v, want 0", got.Slope)
		}
		if !almostEqual(got.Forecast, 42) {
			t.Errorf("flat series forecast = %v, want 42", got.Forecast)
		}
	})

	t.Run("R2 in [0, 1]", func(t *testing.T) {
		// Random-ish prices should give R² in valid range.
		prices := []float64{10, 12, 8, 15, 11, 14, 9, 16, 13, 10}
		got, _ := LinearRegression(prices)
		if got.R2 < 0 || got.R2 > 1 {
			t.Errorf("R² = %v out of [0,1]", got.R2)
		}
	})

	t.Run("insufficient data", func(t *testing.T) {
		_, err := LinearRegression([]float64{1, 2})
		if !errors.Is(err, ErrInsufficientData) {
			t.Errorf("expected ErrInsufficientData, got %v", err)
		}
	})
}

// ── emaFull internal ──────────────────────────────────────────────────────────

func TestEmaFull(t *testing.T) {
	t.Run("length preserved", func(t *testing.T) {
		prices := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		series := emaFull(prices, 3)
		if len(series) != len(prices) {
			t.Errorf("emaFull length: got %d, want %d", len(series), len(prices))
		}
	})

	t.Run("seed value equals SMA", func(t *testing.T) {
		prices := []float64{10, 20, 30, 40, 50}
		period := 3
		series := emaFull(prices, period)
		sma, _ := SMA(prices[:period], period)
		if !almostEqual(series[period-1], sma) {
			t.Errorf("seed at [period-1]: got %v, want SMA %v", series[period-1], sma)
		}
	})
}
