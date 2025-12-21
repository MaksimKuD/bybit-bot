package bybit

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"bybit-bot/internal/config"
	"bybit-bot/internal/exchange"
)

type Client struct {
	apiKey    string
	apiSecret string
	baseURL   string
	http      *http.Client
}

func New(cfg config.Bybit) *Client {
	baseURL := "https://api.bybit.com"
	if cfg.Testnet {
		baseURL = "https://api-testnet.bybit.com"
	}

	return &Client{
		apiKey:    cfg.ApiKey,
		apiSecret: cfg.ApiSecret,
		baseURL:   baseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// timestamp + api_key + recv_window + query_string
func (c *Client) sign(payload string) string {
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}
func (c *Client) doRequest(method, endpoint string, query map[string]string) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	recvWindow := "5000"

	var buf bytes.Buffer
	first := true
	for k, v := range query {
		if !first {
			buf.WriteString("&")
		}
		first = false
		buf.WriteString(k + "=" + v)
	}

	queryString := buf.String()

	payload := timestamp + c.apiKey + recvWindow + queryString
	signature := c.sign(payload)

	req, err := http.NewRequest(method, c.baseURL+endpoint+"?"+queryString, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-BAPI-API-KEY", c.apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", timestamp)
	req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)
	req.Header.Set("X-BAPI-SIGN", signature)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
func (c *Client) GetBalance() (*exchange.Balance, error) {
	body, err := c.doRequest(
		"GET",
		"/v5/account/wallet-balance",
		map[string]string{
			"accountType": "UNIFIED",
		},
	)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result struct {
			List []struct {
				Coin []struct {
					Coin      string `json:"coin"`
					Available string `json:"availableToWithdraw"`
				} `json:"coin"`
			} `json:"list"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	for _, c := range resp.Result.List[0].Coin {
		if c.Coin == "USDT" {
			val, _ := strconv.ParseFloat(c.Available, 64)
			return &exchange.Balance{
				Currency:  "USDT",
				Available: val,
			}, nil
		}
	}

	return nil, fmt.Errorf("USDT balance not found")
}
func (c *Client) GetPosition(symbol string) (*exchange.Position, error) {
	body, err := c.doRequest(
		"GET",
		"/v5/position/list",
		map[string]string{
			"category": "linear",
			"symbol":   symbol,
		},
	)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result struct {
			List []struct {
				Symbol string `json:"symbol"`
				Side   string `json:"side"`
				Size   string `json:"size"`
				Entry  string `json:"avgPrice"`
			} `json:"list"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	if len(resp.Result.List) == 0 {
		return &exchange.Position{Symbol: symbol, Side: "None"}, nil
	}

	p := resp.Result.List[0]
	size, _ := strconv.ParseFloat(p.Size, 64)
	entry, _ := strconv.ParseFloat(p.Entry, 64)

	return &exchange.Position{
		Symbol: p.Symbol,
		Side:   p.Side,
		Size:   size,
		Entry:  entry,
	}, nil
}
