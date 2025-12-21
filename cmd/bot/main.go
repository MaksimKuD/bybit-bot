package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bybit-bot/internal/app"
	"bybit-bot/internal/executor"
	mdbybit "bybit-bot/internal/marketdata/bybit"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Stopping bot...")
		cancel()
	}()

	// === зависимости ===
	md := mdbybit.NewDemo()            // market data
	exec := executor.NewMockExecutor() // ⛔ пока mock

	app := app.New(
		md,
		exec,
		"BTCUSDT",
		5*time.Minute,
	)

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
