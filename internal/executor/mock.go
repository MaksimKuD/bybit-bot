package executor

import (
	"log"

	"bybit-bot/internal/strategy"
)

type MockExecutor struct {
	side          PositionSide
	entryPrice    float64
	unrealizedPnL float64
	realizedPnL   float64
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{side: SideNone}
}

func (e *MockExecutor) OnSignal(signal strategy.Signal, price float64) error {

	switch signal {

	case strategy.BUY:
		if e.side == SideNone {
			e.side = SideLong
			e.entryPrice = price
			log.Println("MOCK: OPEN LONG @", price)
		}

	case strategy.SELL:
		if e.side == SideNone {
			e.side = SideShort
			e.entryPrice = price
			log.Println("MOCK: OPEN SHORT @", price)
		}

	case strategy.EXIT:
		if e.side == SideLong {
			pnl := price - e.entryPrice
			e.realizedPnL += pnl
			e.side = SideNone
			log.Println("MOCK: CLOSE LONG @", price, "PnL:", pnl)
		}

		if e.side == SideShort {
			pnl := e.entryPrice - price
			e.realizedPnL += pnl
			e.side = SideNone
			log.Println("MOCK: CLOSE SHORT @", price, "PnL:", pnl)
		}
	}

	return nil
}

func (e *MockExecutor) State() Position {
	return Position{
		Side:          e.side,
		EntryPrice:    e.entryPrice,
		UnrealizedPnL: e.unrealizedPnL,
		RealizedPnL:   e.realizedPnL,
	}
}
