package executor

import "bybit-bot/internal/strategy"

type Executor interface {
	OnSignal(signal strategy.Signal, price float64) error
	State() Position
}
