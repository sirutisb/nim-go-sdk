package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
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
// Tries multiple sources in order of preference
func getEthereumGasFees() (map[string]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// Try method 1: Direct RPC with eth_gasPrice only (most reliable)
	result, err := getEthGasViaRPC(client)
	if err == nil {
		return result, nil
	}

	// Try method 2: Blocknative API (no key required)
	result, err = getEthGasViaBlocknative(client)
	if err == nil {
		return result, nil
	}

	return nil, fmt.Errorf("all Ethereum gas API sources failed")
}

// getEthGasViaRPC uses direct Ethereum JSON-RPC calls
func getEthGasViaRPC(client *http.Client) (map[string]interface{}, error) {
	rpcURL := "https://ethereum-rpc.publicnode.com"

	// Make eth_gasPrice call
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_gasPrice",
		"params":  []interface{}{},
		"id":      1,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", rpcURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse as generic map to handle various response formats
	var rpcResp map[string]interface{}
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	// Check for error
	if errObj, ok := rpcResp["error"]; ok && errObj != nil {
		return nil, fmt.Errorf("RPC error: %v", errObj)
	}

	// Get result - it should be a hex string
	result, ok := rpcResp["result"]
	if !ok || result == nil {
		return nil, fmt.Errorf("no result in RPC response")
	}

	resultStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("result is not a string: %T", result)
	}

	gasPrice, err := hexToGwei(resultStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert gas price: %w", err)
	}

	return buildEthResponse(client, gasPrice, gasPrice*0.8, gasPrice, gasPrice*1.2)
}

// getEthGasViaBlocknative uses the Blocknative Gas API
func getEthGasViaBlocknative(client *http.Client) (map[string]interface{}, error) {
	gasURL := "https://api.blocknative.com/gasprices/blockprices?chainid=1"

	resp, err := client.Get(gasURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Blocknative: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Blocknative returned status: %d", resp.StatusCode)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Blocknative response: %w", err)
	}

	// Parse as generic map for flexibility
	var data map[string]interface{}
	if err := json.Unmarshal(respBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode Blocknative response: %w", err)
	}

	// Extract blockPrices array
	blockPrices, ok := data["blockPrices"].([]interface{})
	if !ok || len(blockPrices) == 0 {
		return nil, fmt.Errorf("no blockPrices in Blocknative response")
	}

	// Get first block price
	firstBlock, ok := blockPrices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid blockPrices format")
	}

	// Get base fee
	baseFee, _ := toFloat64(firstBlock["baseFeePerGas"])

	// Get estimated prices
	estimatedPrices, ok := firstBlock["estimatedPrices"].([]interface{})
	if !ok || len(estimatedPrices) == 0 {
		return nil, fmt.Errorf("no estimatedPrices in Blocknative response")
	}

	// Extract prices at different confidence levels
	var safeGas, proposeGas, fastGas float64
	for _, ep := range estimatedPrices {
		price, ok := ep.(map[string]interface{})
		if !ok {
			continue
		}
		confidence, _ := toFloat64(price["confidence"])
		maxFee, _ := toFloat64(price["maxFeePerGas"])

		switch int(confidence) {
		case 99:
			fastGas = maxFee
		case 90:
			proposeGas = maxFee
		case 70:
			safeGas = maxFee
		}
	}

	// Fallbacks
	if fastGas == 0 {
		firstPrice, ok := estimatedPrices[0].(map[string]interface{})
		if ok {
			fastGas, _ = toFloat64(firstPrice["maxFeePerGas"])
		}
	}
	if proposeGas == 0 {
		proposeGas = fastGas * 0.9
	}
	if safeGas == 0 {
		safeGas = fastGas * 0.7
	}
	if baseFee == 0 {
		baseFee = proposeGas
	}

	return buildEthResponse(client, baseFee, safeGas, proposeGas, fastGas)
}

// buildEthResponse creates the final response map for Ethereum
func buildEthResponse(client *http.Client, baseFee, safeGas, proposeGas, fastGas float64) (map[string]interface{}, error) {
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

	// Calculate estimated transaction cost (21000 gas for simple ETH transfer)
	gasLimit := 21000
	gweiPrice := proposeGas
	weiCost := gweiPrice * float64(gasLimit) * 1e9
	ethCost := weiCost / 1e18
	usdCost := ethCost * ethPrice

	// Determine traffic level based on base fee
	trafficLevel := getTrafficLevel(baseFee)

	return map[string]interface{}{
		"blockchain":    "Ethereum",
		"unit":          "gwei",
		"current_price": fmt.Sprintf("$%.2f", ethPrice),
		"fees": map[string]interface{}{
			"safe":     fmt.Sprintf("%.2f", safeGas),
			"standard": fmt.Sprintf("%.2f", proposeGas),
			"fast":     fmt.Sprintf("%.2f", fastGas),
			"base_fee": fmt.Sprintf("%.6f", baseFee),
		},
		"estimated_tx_cost_usd": fmt.Sprintf("$%.4f", usdCost),
		"traffic_level":         trafficLevel,
		"recommendation":        getRecommendation(trafficLevel),
	}, nil
}

// hexToGwei converts a hex string (wei) to gwei as float64
func hexToGwei(hexStr string) (float64, error) {
	// Remove 0x prefix if present
	hexStr = strings.TrimPrefix(hexStr, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")

	if hexStr == "" {
		return 0, fmt.Errorf("empty hex string")
	}

	wei := new(big.Int)
	_, success := wei.SetString(hexStr, 16)
	if !success {
		return 0, fmt.Errorf("invalid hex value: %s", hexStr)
	}

	// Convert wei to gwei (divide by 1e9)
	gwei := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e9)
	gwei.Quo(gwei, divisor)

	result, _ := gwei.Float64()
	return result, nil
}

// toFloat64 safely converts interface{} to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
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
