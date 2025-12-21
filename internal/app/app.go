package app

import (
	"context"
	"log"
	"time"

	"bybit-bot/internal/executor"
	"bybit-bot/internal/marketdata"
	"bybit-bot/internal/strategy"
	"bybit-bot/internal/strategy/ema_rsi_atr"
)

type App struct {
	md       marketdata.MarketData
	exec     executor.Executor
	strategy strategy.Strategy

	lastCandleTime time.Time
	timeframe      time.Duration
	symbol         string
}

func New(
	md marketdata.MarketData,
	exec executor.Executor,
	symbol string,
	tf time.Duration,
) *App {
	return &App{
		md:        md,
		exec:      exec,
		strategy:  ema_rsi_atr.New(),
		symbol:    symbol,
		timeframe: tf,
	}
}

func (a *App) Run(ctx context.Context) error {

	log.Println("Starting trading bot...")
	log.Println("Symbol:", a.symbol)
	log.Println("Timeframe:", a.timeframe)

	ticker := time.NewTicker(a.timeframe)
	defer ticker.Stop()

	// первый запуск сразу
	a.processOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopped")
			return nil

		case <-ticker.C:
			a.processOnce(ctx)
		}
	}
}

func (a *App) processOnce(ctx context.Context) {

	candles, err := a.md.GetCandles(a.symbol, "5m", 2)
	if err != nil {
		log.Println("MarketData error:", err)
		return
	}

	c := candles[len(candles)-1]

	// 🛡 анти-дубликат
	if c.Time.Equal(a.lastCandleTime) {
		return
	}
	a.lastCandleTime = c.Time

	log.Printf(
		"CANDLE %s O=%.2f H=%.2f L=%.2f C=%.2f",
		c.Time.Format(time.RFC3339),
		c.Open, c.High, c.Low, c.Close,
	)

	signal := a.strategy.OnCandle(strategy.Candle(c))
	log.Println("Strategy signal:", signal)

	if err := a.exec.OnSignal(signal, c.Close); err != nil {
		log.Println("Executor error:", err)
	}

	pos := a.exec.State()
	log.Printf(
		"POS=%s Entry=%.2f PnL=%.2f Realized=%.2f",
		pos.Side,
		pos.EntryPrice,
		pos.UnrealizedPnL,
		pos.RealizedPnL,
	)
}
