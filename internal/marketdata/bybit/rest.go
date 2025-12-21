package bybit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bybit-bot/internal/marketdata"
)

// REST — клиент Bybit Market Data (Demo / Mainnet)
type REST struct {
	baseURL string
	http    *http.Client
}

// NewDemo создаёт REST клиент для Demo / публичного REST
func NewDemo() *REST {
	return &REST{
		baseURL: "https://api.bybit.com",
		http: &http.Client{
			Timeout: 30 * time.Second, // увеличенный таймаут
		},
	}
}

// GetCandles возвращает свечи с retry
func (r *REST) GetCandles(symbol, timeframe string, limit int) ([]marketdata.Candle, error) {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		candles, err := r.getCandlesOnce(symbol, timeframe, limit)
		if err == nil {
			return candles, nil
		}

		lastErr = err
		time.Sleep(time.Duration(attempt) * time.Second) // экспоненциальная пауза
	}

	return nil, lastErr
}

// getCandlesOnce делает один запрос к Bybit
func (r *REST) getCandlesOnce(symbol, timeframe string, limit int) ([]marketdata.Candle, error) {
	req, err := http.NewRequest(
		"GET",
		r.baseURL+"/v5/market/kline",
		nil,
	)
	if err != nil {
		return nil, err
	}

	// ВАЖНО: User-Agent помогает избежать блокировок
	req.Header.Set("User-Agent", "bybit-bot/1.0")

	q := req.URL.Query()
	q.Add("category", "linear")
	q.Add("symbol", symbol)
	q.Add("interval", normalizeTF(timeframe))
	q.Add("limit", strconv.Itoa(limit))
	req.URL.RawQuery = q.Encode()

	resp, err := r.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Result struct {
			List [][]string `json:"list"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	candles := make([]marketdata.Candle, 0, len(raw.Result.List))

	for i := len(raw.Result.List) - 1; i >= 0; i-- {
		k := raw.Result.List[i]

		ts, _ := strconv.ParseInt(k[0], 10, 64)
		open, _ := strconv.ParseFloat(k[1], 64)
		high, _ := strconv.ParseFloat(k[2], 64)
		low, _ := strconv.ParseFloat(k[3], 64)
		closep, _ := strconv.ParseFloat(k[4], 64)
		vol, _ := strconv.ParseFloat(k[5], 64)

		candles = append(candles, marketdata.Candle{
			Time:   time.UnixMilli(ts),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closep,
			Volume: vol,
		})
	}

	if len(candles) == 0 {
		return nil, fmt.Errorf("no candles returned")
	}

	return candles, nil
}

// normalizeTF переводит "5m" → "5" для Bybit
func normalizeTF(tf string) string {
	switch tf {
	case "1m":
		return "1"
	case "5m":
		return "5"
	case "15m":
		return "15"
	case "30m":
		return "30"
	case "1h":
		return "60"
	case "4h":
		return "240"
	case "1d":
		return "D"
	default:
		return "5"
	}
}
