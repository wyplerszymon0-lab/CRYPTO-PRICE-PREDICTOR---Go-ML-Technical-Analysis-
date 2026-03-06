* Crypto Price Predictor
A real-time cryptocurrency price prediction engine built in Go — fetches live market data from CoinGecko and applies technical analysis algorithms to generate actionable buy/sell signals.

* Features

* Live data — fetches real closing prices from CoinGecko API (no API key required)
* 4 technical indicators — RSI, MACD, Bollinger Bands, SMA crossover
* ML-style scoring engine — weighted composite score from -1.0 to +1.0
* Price prediction — next-day price forecast via linear regression
* Colored terminal output — clean, readable reports per coin
*  Zero dependencies — pure Go standard library only


* Demo Output
╔══════════════════════════════════════╗
║     CRYPTO PRICE PREDICTOR v1.0      ║
║     Go · ML · Technical Analysis     ║
╚══════════════════════════════════════╝

* Fetching data for bitcoin...

═══ BITCOIN ═══
  Current Price:       $67,420.00
  Predicted (next day): $68,150.32
  Expected Change:     +1.08%

  Signal:              STRONG BUY
  Confidence:          74.2%

  ── Indicators ──
  RSI (14):            28.41  (Oversold)
  MACD Line:           124.3200
  Bollinger Position:  18.4%
  SMA Trend:           ↑ BULLISH
  Composite Score:     0.742

*Getting Started
Prerequisites

*Go 1.22+
Internet connection (for live CoinGecko data)

*Installation & Run
bashgit clone https://github.com/wyplerszymon0-lab/crypto-predictor.git
cd crypto-predictor
go run main.go
That's it — no go get, no external packages.

* How It Works
The engine combines 4 technical indicators into a single weighted score:
IndicatorWeightLogicRSI (14)25%< 30 → oversold (bullish), > 70 → overbought (bearish)MACD25%Histogram + line directionBollinger Bands20%Price position within upper/lower bandsSMA 7/20 crossover30%Golden cross (bullish) / Death cross (bearish)
The composite score maps to signals:
ScoreSignal≥ 0.6🟢 STRONG BUY0.2 – 0.6🟢 BUY-0.2 – 0.2🟡 NEUTRAL-0.6 – -0.2🔴 SELL≤ -0.6🔴 STRONG SELL
Price prediction uses linear regression on the last 7 days of closing prices to extrapolate the next-day value.

* Project Structure
crypto-predictor/
├── main.go                        # Entry point
├── go.mod                         # Module definition
└── internal/
    ├── api/
    │   └── coingecko.go           # CoinGecko API client
    ├── indicators/
    │   └── indicators.go          # SMA, EMA, RSI, MACD, Bollinger Bands
    ├── predictor/
    │   └── engine.go              # Scoring engine + linear regression
    └── report/
        └── report.go              # Colored terminal output

* Supported Coins
By default the engine analyses Bitcoin, Ethereum and Solana. To add more coins, edit the symbols slice in main.go:
gosymbols := []string{"bitcoin", "ethereum", "solana", "cardano", "dogecoin"}
Any coin ID from CoinGecko works.

* Disclaimer
This project is for educational purposes only. It is not financial advice. Cryptocurrency markets are highly volatile — never make investment decisions based solely on automated signals.

* Author
Szymon — @wyplerszymon0-lab

* License
MIT License — feel free to use, modify and
