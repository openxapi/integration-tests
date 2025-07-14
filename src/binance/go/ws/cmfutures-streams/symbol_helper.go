package streamstest

import (
	"context"
	"fmt"
	"strings"
	"testing"

	cmfutures "github.com/openxapi/binance-go/rest/cmfutures"
)


// getBTCSymbols returns BTC-related symbols from the REST API
func getBTCSymbols(t *testing.T) ([]string, error) {
	testSymbols, err := getTestSymbols(t)
	if err != nil {
		return nil, err
	}
	
	var btcSymbols []string
	if testSymbols["btc_perp"] != "" {
		btcSymbols = append(btcSymbols, testSymbols["btc_perp"])
	}
	if testSymbols["btc_pair"] != "" {
		btcSymbols = append(btcSymbols, testSymbols["btc_pair"])
	}
	
	return btcSymbols, nil
}

// getTestSymbols fetches valid symbols from the REST API
func getTestSymbols(t *testing.T) (map[string]string, error) {
	// Create a new client for cmfutures REST API
	config := cmfutures.NewConfiguration()
	config.Host = "testnet.binancefuture.com"
	config.Scheme = "https"
	
	client := cmfutures.NewAPIClient(config)

	// Get exchange info
	ctx := context.Background()
	resp, _, err := client.FuturesAPI.GetExchangeInfoV1(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("error getting exchange info: %v", err)
	}

	testSymbols := map[string]string{
		"btc_perp":   "",
		"link_perp":  "",
		"ada_perp":   "",
		"bnb_perp":   "",
		"btc_pair":   "",
		"link_pair":  "",
	}
	
	// Find perpetual contracts and pairs from the API response
	for _, symbol := range resp.Symbols {
		if symbol.Symbol != nil && symbol.ContractStatus != nil && *symbol.ContractStatus == "TRADING" {
			symbolName := strings.ToLower(*symbol.Symbol)
			symbolUpper := strings.ToUpper(*symbol.Symbol)
			
			// Look for perpetual contracts
			if symbol.ContractType != nil && *symbol.ContractType == "PERPETUAL" {
				if strings.Contains(symbolUpper, "BTCUSD") && testSymbols["btc_perp"] == "" {
					testSymbols["btc_perp"] = symbolName
				} else if strings.Contains(symbolUpper, "LINKUSD") && testSymbols["link_perp"] == "" {
					testSymbols["link_perp"] = symbolName
				} else if strings.Contains(symbolUpper, "ADAUSD") && testSymbols["ada_perp"] == "" {
					testSymbols["ada_perp"] = symbolName
				} else if strings.Contains(symbolUpper, "BNBUSD") && testSymbols["bnb_perp"] == "" {
					testSymbols["bnb_perp"] = symbolName
				}
			}
			
			// Look for pairs for index price streams (use the pair field)
			if symbol.Pair != nil {
				pairName := strings.ToLower(*symbol.Pair)
				pairUpper := strings.ToUpper(*symbol.Pair)
				if strings.Contains(pairUpper, "BTCUSD") && testSymbols["btc_pair"] == "" {
					testSymbols["btc_pair"] = pairName
				} else if strings.Contains(pairUpper, "LINKUSD") && testSymbols["link_pair"] == "" {
					testSymbols["link_pair"] = pairName
				}
			}
		}
	}
	
	t.Logf("Found symbols from REST API: %+v", testSymbols)
	return testSymbols, nil
}

// logAvailableSymbols logs available symbols from the REST API
func logAvailableSymbols(t *testing.T) {
	symbols, err := getTestSymbols(t)
	if err != nil {
		t.Logf("Error getting symbols: %v", err)
		return
	}
	
	t.Logf("=== Available Coin-M Futures Symbols (from REST API) ===")
	for key, symbol := range symbols {
		if symbol != "" {
			t.Logf("%s: %s", key, symbol)
		} else {
			t.Logf("%s: NOT FOUND", key)
		}
	}
	t.Logf("Note: Stream names use lowercase, event symbols are uppercase")
}