package executor

import (
	"log"

	"bybit-bot/internal/strategy"
)

type BybitExecutor struct {
	side PositionSide
}

func NewBybitExecutor() *BybitExecutor {
	return &BybitExecutor{
		side: SideNone,
	}
}

// === executor interface ===

func (e *BybitExecutor) OnSignal(signal strategy.Signal, price float64) error {

	// ⚠️ ШАГ 10.1:
	// Пока НИЧЕГО не торгуем
	// Только логируем сигналы

	switch signal {
	case strategy.BUY:
		log.Println("BYBIT EXECUTOR: BUY signal received (not executed)")
	case strategy.SELL:
		log.Println("BYBIT EXECUTOR: SELL signal received (not executed)")
	}

	return nil
}

func (e *BybitExecutor) State() Position {
	return Position{
		Side: e.side,
	}
}
