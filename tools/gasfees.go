package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
)

// Gas fee API response structures
type MempoolFeesResponse struct {
	FastestFee  int `json:"fastestFee"`
	HalfHourFee int `json:"halfHourFee"`
	HourFee     int `json:"hourFee"`
	EconomyFee  int `json:"economyFee"`
	MinimumFee  int `json:"minimumFee"`
}

type EtherscanGasResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		SafeGasPrice    string `json:"SafeGasPrice"`
		ProposeGasPrice string `json:"ProposeGasPrice"`
		FastGasPrice    string `json:"FastGasPrice"`
		SuggestBaseFee  string `json:"suggestBaseFee"`
	} `json:"result"`
}

type CoinbasePriceResponse struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

// NewGasFeeTool creates a tool for checking blockchain gas/transaction fees
func NewGasFeeTool() core.Tool {
	return New("get_gas_fees").
		Description("Get current blockchain transaction fees and estimated costs for Bitcoin or Ethereum. Returns fee rates, price, and estimated transaction cost in USD.").
		Schema(ObjectSchema(map[string]interface{}{
			"blockchain": StringEnumProperty("Blockchain to check fees for", "bitcoin", "btc", "ethereum", "eth"),
		}, "blockchain")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				Blockchain string `json:"blockchain"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			switch params.Blockchain {
			case "bitcoin", "btc":
				return getBitcoinGasFees()
			case "ethereum", "eth":
				return getEthereumGasFees()
			default:
				return nil, fmt.Errorf("unsupported blockchain: %s (use 'bitcoin' or 'ethereum')", params.Blockchain)
			}
		}).
		Build()
}

// getBitcoinGasFees fetches current Bitcoin transaction fees
func getBitcoinGasFees() (map[string]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Fetch fee data from mempool.space
	feeResp, err := client.Get("https://mempool.space/api/v1/fees/recommended")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Bitcoin fees: %w", err)
	}
	defer feeResp.Body.Close()

	if feeResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bitcoin fee API returned status: %d", feeResp.StatusCode)
	}

	var fees MempoolFeesResponse
	if err := json.NewDecoder(feeResp.Body).Decode(&fees); err != nil {
		return nil, fmt.Errorf("failed to decode fee response: %w", err)
	}

	// Fetch BTC price
	priceResp, err := client.Get("https://api.coinbase.com/v2/prices/BTC-USD/spot")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BTC price: %w", err)
	}
	defer priceResp.Body.Close()

	var priceData CoinbasePriceResponse
	if err := json.NewDecoder(priceResp.Body).Decode(&priceData); err != nil {
		return nil, fmt.Errorf("failed to decode price response: %w", err)
	}

	btcPrice, _ := strconv.ParseFloat(priceData.Data.Amount, 64)

	// Calculate estimated transaction cost (assuming 140 vBytes typical tx)
	txSizeVBytes := 140
	satsForTx := float64(fees.HourFee * txSizeVBytes)
	btcForTx := satsForTx / 100_000_000
	usdCost := btcForTx * btcPrice

	// Determine traffic level
	trafficLevel := getTrafficLevel(float64(fees.HourFee))

	return map[string]interface{}{
		"blockchain":    "Bitcoin",
		"unit":          "sat/vB",
		"current_price": fmt.Sprintf("$%.2f", btcPrice),
		"fees": map[string]interface{}{
			"fastest":   fees.FastestFee,
			"half_hour": fees.HalfHourFee,
			"hour":      fees.HourFee,
			"economy":   fees.EconomyFee,
			"minimum":   fees.MinimumFee,
		},
		"estimated_tx_cost_usd": fmt.Sprintf("$%.4f", usdCost),
		"traffic_level":         trafficLevel,
		"recommendation":        getRecommendation(trafficLevel),
	}, nil
}

// getEthereumGasFees fetches current Ethereum gas prices
func getEthereumGasFees() (map[string]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Fetch gas data from Etherscan (free tier, no API key required for basic calls)
	// Using public API - for production, add your own API key
	gasURL := "https://api.etherscan.io/api?module=gastracker&action=gasoracle"

	gasResp, err := client.Get(gasURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Ethereum gas prices: %w", err)
	}
	defer gasResp.Body.Close()

	if gasResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Etherscan API returned status: %d", gasResp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(gasResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gasData EtherscanGasResponse
	if err := json.Unmarshal(bodyBytes, &gasData); err != nil {
		return nil, fmt.Errorf("failed to decode gas response: %w", err)
	}

	if gasData.Status != "1" {
		return nil, fmt.Errorf("Etherscan API error: %s", gasData.Message)
	}

	// Fetch ETH price
	priceResp, err := client.Get("https://api.coinbase.com/v2/prices/ETH-USD/spot")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ETH price: %w", err)
	}
	defer priceResp.Body.Close()

	var priceData CoinbasePriceResponse
	if err := json.NewDecoder(priceResp.Body).Decode(&priceData); err != nil {
		return nil, fmt.Errorf("failed to decode price response: %w", err)
	}

	ethPrice, _ := strconv.ParseFloat(priceData.Data.Amount, 64)

	// Parse gas prices
	safeGas, _ := strconv.ParseFloat(gasData.Result.SafeGasPrice, 64)
	proposeGas, _ := strconv.ParseFloat(gasData.Result.ProposeGasPrice, 64)
	fastGas, _ := strconv.ParseFloat(gasData.Result.FastGasPrice, 64)

	// Calculate estimated transaction cost (21000 gas for simple ETH transfer)
	gasLimit := 21000
	gweiPrice := proposeGas
	weiCost := gweiPrice * float64(gasLimit) * 1e9
	ethCost := weiCost / 1e18
	usdCost := ethCost * ethPrice

	// Determine traffic level
	trafficLevel := getTrafficLevel(proposeGas)

	return map[string]interface{}{
		"blockchain":    "Ethereum",
		"unit":          "gwei",
		"current_price": fmt.Sprintf("$%.2f", ethPrice),
		"fees": map[string]interface{}{
			"safe":     safeGas,
			"standard": proposeGas,
			"fast":     fastGas,
			"base_fee": gasData.Result.SuggestBaseFee,
		},
		"estimated_tx_cost_usd": fmt.Sprintf("$%.4f", usdCost),
		"traffic_level":         trafficLevel,
		"recommendation":        getRecommendation(trafficLevel),
	}, nil
}

// getTrafficLevel determines network congestion level
func getTrafficLevel(fee float64) string {
	switch {
	case fee <= 5:
		return "LOW"
	case fee <= 20:
		return "MEDIUM"
	case fee <= 50:
		return "HIGH"
	default:
		return "VERY HIGH"
	}
}

// getRecommendation provides advice based on traffic level
func getRecommendation(level string) string {
	switch level {
	case "LOW":
		return "Great time to transact! Fees are low."
	case "MEDIUM":
		return "Moderate fees. Good for non-urgent transactions."
	case "HIGH":
		return "Network is busy. Consider waiting if not urgent."
	case "VERY HIGH":
		return "Network is congested. Wait for lower fees if possible."
	default:
		return "Check current conditions before transacting."
	}
}
