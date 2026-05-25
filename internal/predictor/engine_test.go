package predictor

import (
	"testing"
)

// makePrices returns n prices following a linear trend starting at start with given slope.
func makePrices(n int, start, slope float64) []float64 {
	prices := make([]float64, n)
	for i := range prices {
		prices[i] = start + slope*float64(i)
	}
	return prices
}

func TestEngine_Predict(t *testing.T) {
	t.Run("returns result with enough data", func(t *testing.T) {
		prices := makePrices(60, 100, 0.5)
		eng := NewEngine(prices)
		result, err := eng.Predict()
		if err != nil {
			t.Fatalf("Predict() error: %v", err)
		}
		if result.CurrentPrice != prices[len(prices)-1] {
			t.Errorf("CurrentPrice = %v, want %v", result.CurrentPrice, prices[len(prices)-1])
		}
		if result.Confidence < 0 || result.Confidence > 100 {
			t.Errorf("Confidence = %v out of [0,100]", result.Confidence)
		}
		if result.ForecastR2 < 0 || result.ForecastR2 > 1 {
			t.Errorf("ForecastR2 = %v out of [0,1]", result.ForecastR2)
		}
	})

	t.Run("insufficient data returns error", func(t *testing.T) {
		eng := NewEngine([]float64{1, 2, 3})
		_, err := eng.Predict()
		if err == nil {
			t.Error("expected error for insufficient data, got nil")
		}
	})

	t.Run("score is in [-1, 1]", func(t *testing.T) {
		prices := makePrices(90, 100, 0.5)
		eng := NewEngine(prices)
		result, err := eng.Predict()
		if err != nil {
			t.Fatal(err)
		}
		if result.Score < -1 || result.Score > 1 {
			t.Errorf("Score = %v out of [-1,1]", result.Score)
		}
	})

	t.Run("confidence matches absolute score", func(t *testing.T) {
		prices := makePrices(90, 100, 0.5)
		eng := NewEngine(prices)
		result, _ := eng.Predict()
		want := result.Score * 100
		if want < 0 {
			want = -want
		}
		if want > 100 {
			want = 100
		}
		if result.Confidence != want {
			t.Errorf("Confidence = %v, want %v", result.Confidence, want)
		}
	})

	t.Run("signal matches score threshold", func(t *testing.T) {
		prices := makePrices(90, 100, 0.5)
		eng := NewEngine(prices)
		result, _ := eng.Predict()
		sig := scoreToSignal(result.Score)
		if sig != result.Signal {
			t.Errorf("Signal = %v, expected %v for score %v", result.Signal, sig, result.Score)
		}
	})
}

func TestEngine_PredictAt(t *testing.T) {
	prices := makePrices(60, 100, 0.5)
	eng := NewEngine(prices)

	t.Run("PredictAt(i) matches sub-slice Predict()", func(t *testing.T) {
		i := 45
		r1, err1 := eng.PredictAt(i)
		r2, err2 := NewEngine(prices[:i+1]).Predict()

		if (err1 != nil) != (err2 != nil) {
			t.Fatalf("PredictAt error mismatch: %v vs %v", err1, err2)
		}
		if err1 != nil {
			return
		}
		if r1.Score != r2.Score {
			t.Errorf("Score mismatch: PredictAt=%v sub-slice=%v", r1.Score, r2.Score)
		}
		if r1.CurrentPrice != r2.CurrentPrice {
			t.Errorf("CurrentPrice mismatch: %v vs %v", r1.CurrentPrice, r2.CurrentPrice)
		}
	})

	t.Run("out-of-range index returns error", func(t *testing.T) {
		_, err := eng.PredictAt(len(prices))
		if err == nil {
			t.Error("expected error for out-of-range index")
		}
	})
}

func TestSignal_String(t *testing.T) {
	tests := []struct {
		signal Signal
		want   string
	}{
		{StrongBuy, "STRONG BUY"},
		{Buy, "BUY"},
		{Neutral, "NEUTRAL"},
		{Sell, "SELL"},
		{StrongSell, "STRONG SELL"},
	}
	for _, tt := range tests {
		if got := tt.signal.String(); got != tt.want {
			t.Errorf("Signal(%d).String() = %q, want %q", int(tt.signal), got, tt.want)
		}
	}
}

func TestSignal_Ordering(t *testing.T) {
	// The backtest engine relies on this ordering for comparisons.
	if !(StrongSell < Sell && Sell < Neutral && Neutral < Buy && Buy < StrongBuy) {
		t.Error("signal ordering invariant violated")
	}
	if Buy < Neutral || StrongBuy < Buy {
		t.Error("bullish signals must be > Neutral")
	}
	if Sell > Neutral || StrongSell > Sell {
		t.Error("bearish signals must be < Neutral")
	}
}
