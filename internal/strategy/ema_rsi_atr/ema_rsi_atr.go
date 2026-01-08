package ema_rsi_atr

import (
	"math"

	"bybit-bot/internal/strategy"
)

// ================== CONFIG ==================

const (
	emaLength = 100
	rsiLength = 14
	atrLength = 14

	atrMultSL        = 2.0
	rrRatioFinal     = 3.0
	beActivationMult = 2.5
)

// ================== STATE ==================

type Strategy struct {
	candles []strategy.Candle

	inPosition    bool
	positionSide  strategy.Signal
	entryPrice    float64
	stopLoss      float64
	takeProfit    float64
	isBEActivated bool
}

func New() *Strategy {
	return &Strategy{
		candles: make([]strategy.Candle, 0, 300),
	}
}

// ================== MAIN ENTRY ==================

func (s *Strategy) OnCandle(c strategy.Candle) strategy.Signal {

	// 1️⃣ сохраняем свечу
	s.candles = append(s.candles, c)
	if len(s.candles) > 300 {
		s.candles = s.candles[1:]
	}

	// 2️⃣ проверка на минимальное количество данных
	if len(s.candles) < emaLength+rsiLength+atrLength {
		return strategy.HOLD
	}

	// 3️⃣ расчёт индикаторов
	closes := extractCloses(s.candles)

	ema := calculateEMA(closes, emaLength)
	rsi := calculateRSI(closes, rsiLength)
	atr := calculateATR(s.candles, atrLength)

	i := len(s.candles) - 1

	currentEMA := ema[i]
	currentRSI := rsi[i]
	prevRSI := rsi[i-1]
	currentATR := atr[i]

	// фильтр волатильности
	if currentATR <= 0 || currentATR < c.Close*0.002 {
		return strategy.HOLD
	}

	// 4️⃣ выход
	if s.inPosition {
		if s.checkExit(c, currentATR) {
			return strategy.EXIT
		}
	}

	// 5️⃣ вход
	if !s.inPosition {

		isBullish := c.Close > c.Open
		isBearish := c.Close < c.Open

		longSignal :=
			c.Close > currentEMA &&
				currentRSI > 45 &&
				prevRSI <= 45 &&
				isBullish

		shortSignal :=
			c.Close < currentEMA &&
				currentRSI < 55 &&
				prevRSI >= 55 &&
				isBearish

		if longSignal {
			s.enter(c.Close, currentATR, strategy.BUY)
			return strategy.BUY
		}

		if shortSignal {
			s.enter(c.Close, currentATR, strategy.SELL)
			return strategy.SELL
		}
	}

	return strategy.HOLD
}

// ================== POSITION LOGIC ==================

func (s *Strategy) enter(price, atr float64, side strategy.Signal) {

	slDist := atr * atrMultSL

	if side == strategy.BUY {
		s.stopLoss = price - slDist
		s.takeProfit = price + slDist*rrRatioFinal
	} else {
		s.stopLoss = price + slDist
		s.takeProfit = price - slDist*rrRatioFinal
	}

	s.entryPrice = price
	s.positionSide = side
	s.inPosition = true
	s.isBEActivated = false
}

func (s *Strategy) checkExit(c strategy.Candle, atr float64) bool {

	// TP
	if s.positionSide == strategy.BUY && c.High >= s.takeProfit {
		s.reset()
		return true
	}
	if s.positionSide == strategy.SELL && c.Low <= s.takeProfit {
		s.reset()
		return true
	}

	// BE activation
	if !s.isBEActivated {
		beDist := atr * beActivationMult

		if s.positionSide == strategy.BUY && c.High >= s.entryPrice+beDist {
			s.stopLoss = s.entryPrice + atr*0.3
			s.isBEActivated = true
		}

		if s.positionSide == strategy.SELL && c.Low <= s.entryPrice-beDist {
			s.stopLoss = s.entryPrice - atr*0.3
			s.isBEActivated = true
		}
	}

	// SL / BE
	if s.positionSide == strategy.BUY && c.Low <= s.stopLoss {
		s.reset()
		return true
	}
	if s.positionSide == strategy.SELL && c.High >= s.stopLoss {
		s.reset()
		return true
	}

	return false
}

func (s *Strategy) reset() {
	s.inPosition = false
	s.isBEActivated = false
}

// ================== INDICATORS ==================

func extractCloses(c []strategy.Candle) []float64 {
	out := make([]float64, len(c))
	for i := range c {
		out[i] = c[i].Close
	}
	return out
}

func calculateEMA(closes []float64, period int) []float64 {
	ema := make([]float64, len(closes))
	sum := 0.0

	for i := 0; i < period; i++ {
		sum += closes[i]
	}
	ema[period-1] = sum / float64(period)

	m := 2.0 / (float64(period) + 1.0)

	for i := period; i < len(closes); i++ {
		ema[i] = (closes[i]-ema[i-1])*m + ema[i-1]
	}
	return ema
}

func calculateRSI(closes []float64, period int) []float64 {
	rsi := make([]float64, len(closes))

	up, down := 0.0, 0.0
	for i := 1; i <= period; i++ {
		d := closes[i] - closes[i-1]
		if d > 0 {
			up += d
		} else {
			down -= d
		}
	}

	avgUp := up / float64(period)
	avgDown := down / float64(period)

	if avgDown == 0 {
		rsi[period] = 100
	} else {
		rs := avgUp / avgDown
		rsi[period] = 100 - (100 / (1 + rs))
	}

	for i := period + 1; i < len(closes); i++ {
		d := closes[i] - closes[i-1]
		u, dn := 0.0, 0.0
		if d > 0 {
			u = d
		} else {
			dn = -d
		}
		avgUp = (avgUp*float64(period-1) + u) / float64(period)
		avgDown = (avgDown*float64(period-1) + dn) / float64(period)

		if avgDown == 0 {
			rsi[i] = 100
		} else {
			rs := avgUp / avgDown
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}
	return rsi
}

func calculateATR(c []strategy.Candle, period int) []float64 {
	atr := make([]float64, len(c))

	tr := func(i int) float64 {
		hml := c[i].High - c[i].Low
		hpc := math.Abs(c[i].High - c[i-1].Close)
		lpc := math.Abs(c[i].Low - c[i-1].Close)
		return math.Max(hml, math.Max(hpc, lpc))
	}

	sum := 0.0
	for i := 1; i <= period; i++ {
		sum += tr(i)
	}
	atr[period] = sum / float64(period)

	for i := period + 1; i < len(c); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr(i)) / float64(period)
	}
	return atr
}
