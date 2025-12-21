// internal/strategy/strategy.go
package strategy

import "time"

type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type Signal string

const (
	BUY  Signal = "BUY"
	SELL Signal = "SELL"
	HOLD Signal = "HOLD"
)

type Strategy interface {
	OnCandle(c Candle) Signal
}
