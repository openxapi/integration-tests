package streamstest

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	openapi "github.com/openxapi/binance-go/rest/options"
)

// ActiveSymbolCache caches active symbols to avoid repeated REST API calls
type ActiveSymbolCache struct {
	symbols    []string
	lastUpdate time.Time
	ttl        time.Duration
}

var symbolCache = &ActiveSymbolCache{
	ttl: 5 * time.Minute, // Cache symbols for 5 minutes
}

// getActiveOptionsSymbols returns currently active options symbols using REST API
func getActiveOptionsSymbols(underlying string) ([]string, error) {
	// Check cache first
	if time.Since(symbolCache.lastUpdate) < symbolCache.ttl && len(symbolCache.symbols) > 0 {
		return filterSymbolsByUnderlying(symbolCache.symbols, underlying), nil
	}

	// Setup REST client
	cfg := openapi.NewConfiguration()
	cfg.Servers = openapi.ServerConfigurations{
		{
			URL:         "https://eapi.binance.com",
			Description: "Binance Options API (Production)",
		},
	}

	// Override with custom server if provided
	if serverURL := os.Getenv("BINANCE_OPTIONS_REST_SERVER"); serverURL != "" {
		cfg.Servers[0].URL = serverURL
	}

	client := openapi.NewAPIClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get exchange info
	resp, _, err := client.OptionsAPI.GetExchangeInfoV1(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}

	var activeSymbols []string

	// Extract all active symbols
	if resp.OptionSymbols != nil {
		for _, symbol := range resp.OptionSymbols {
			if symbol.Symbol != nil {
				activeSymbols = append(activeSymbols, *symbol.Symbol)
			}
		}
	}

	// Update cache
	symbolCache.symbols = activeSymbols
	symbolCache.lastUpdate = time.Now()

	return filterSymbolsByUnderlying(activeSymbols, underlying), nil
}

// filterSymbolsByUnderlying filters symbols by underlying asset
func filterSymbolsByUnderlying(symbols []string, underlying string) []string {
	if underlying == "" {
		return symbols
	}

	var filtered []string
	for _, symbol := range symbols {
		if strings.HasPrefix(symbol, underlying+"-") {
			filtered = append(filtered, symbol)
		}
	}
	return filtered
}

// getActiveExpirationDates returns currently active expiration dates for an underlying
func getActiveExpirationDates(underlying string) ([]string, error) {
	symbols, err := getActiveOptionsSymbols(underlying)
	if err != nil {
		return nil, err
	}

	expirationMap := make(map[string]bool)
	var expirations []string

	for _, symbol := range symbols {
		parts := strings.Split(symbol, "-")
		if len(parts) >= 2 {
			expiration := parts[1] // YYMMDD format
			if !expirationMap[expiration] {
				expirationMap[expiration] = true
				expirations = append(expirations, expiration)
			}
		}
	}

	// Sort expirations by date (nearest first)
	sort.Slice(expirations, func(i, j int) bool {
		return expirations[i] < expirations[j]
	})

	return expirations, nil
}

// selectNearestExpirySymbol returns a symbol with the nearest expiration date
func selectNearestExpirySymbol(underlying string, optionType string) (string, error) {
	symbols, err := getActiveOptionsSymbols(underlying)
	if err != nil {
		return "", err
	}

	// Filter by option type if specified (C or P)
	var filteredSymbols []string
	for _, symbol := range symbols {
		if optionType == "" || strings.HasSuffix(symbol, "-"+optionType) {
			filteredSymbols = append(filteredSymbols, symbol)
		}
	}

	if len(filteredSymbols) == 0 {
		return "", fmt.Errorf("no active symbols found for %s %s", underlying, optionType)
	}

	// Sort by expiration date and strike price to get consistent results
	sort.Slice(filteredSymbols, func(i, j int) bool {
		partsI := strings.Split(filteredSymbols[i], "-")
		partsJ := strings.Split(filteredSymbols[j], "-")

		if len(partsI) >= 3 && len(partsJ) >= 3 {
			// Compare expiration dates first
			if partsI[1] != partsJ[1] {
				return partsI[1] < partsJ[1]
			}
			// If same expiration, compare strike prices
			strikeI, _ := strconv.Atoi(partsI[2])
			strikeJ, _ := strconv.Atoi(partsJ[2])
			return strikeI < strikeJ
		}
		return filteredSymbols[i] < filteredSymbols[j]
	})

	return filteredSymbols[0], nil
}

// selectATMSymbol tries to find an at-the-money or near-the-money option
func selectATMSymbol(underlying string, optionType string) (string, error) {
	symbols, err := getActiveOptionsSymbols(underlying)
	if err != nil {
		return "", err
	}

	// Get current underlying price
	currentPrice, err := getCurrentUnderlyingPrice(underlying)
	if err != nil {
		// Fallback to simple middle selection if can't get current price
		return selectNearestExpirySymbol(underlying, optionType)
	}

	// Filter by option type if specified
	var filteredSymbols []string
	for _, symbol := range symbols {
		if optionType == "" || strings.HasSuffix(symbol, "-"+optionType) {
			filteredSymbols = append(filteredSymbols, symbol)
		}
	}

	if len(filteredSymbols) == 0 {
		return "", fmt.Errorf("no active symbols found for %s %s", underlying, optionType)
	}

	// Find the strike price closest to current market price
	var bestSymbol string
	minDifference := float64(999999999) // Large initial value

	for _, symbol := range filteredSymbols {
		parts := strings.Split(symbol, "-")
		if len(parts) >= 3 {
			strikePrice, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				continue
			}
			
			// Calculate absolute difference from current price
			difference := abs(strikePrice - currentPrice)
			if difference < minDifference {
				minDifference = difference
				bestSymbol = symbol
			}
		}
	}

	if bestSymbol == "" {
		return "", fmt.Errorf("no valid strike prices found for %s %s", underlying, optionType)
	}

	return bestSymbol, nil
}

// getPreferredUnderlyingAssets returns the most active underlying assets
func getPreferredUnderlyingAssets() []string {
	// Check environment variable first
	if envUnderlyings := os.Getenv("PREFERRED_UNDERLYING"); envUnderlyings != "" {
		return strings.Split(envUnderlyings, ",")
	}

	// Default to most liquid underlying assets
	return []string{"BTC", "ETH"}
}

// getPreferredOptionType returns the preferred option type for testing
func getPreferredOptionType() string {
	if envType := os.Getenv("PREFERRED_OPTION_TYPE"); envType != "" {
		return envType
	}
	return "C" // Default to Call options
}

// validateSymbolFormat checks if a symbol follows the expected format
func validateSymbolFormat(symbol string) bool {
	parts := strings.Split(symbol, "-")
	if len(parts) != 4 {
		return false
	}

	// Check expiration date format (YYMMDD)
	if len(parts[1]) != 6 {
		return false
	}

	// Check strike price is numeric
	if _, err := strconv.Atoi(parts[2]); err != nil {
		return false
	}

	// Check option type is C or P
	if parts[3] != "C" && parts[3] != "P" {
		return false
	}

	return true
}

// getCurrentUnderlyingPrice gets the current price of the underlying asset
func getCurrentUnderlyingPrice(underlying string) (float64, error) {
	// Setup REST client
	cfg := openapi.NewConfiguration()
	cfg.Servers = openapi.ServerConfigurations{
		{
			URL:         "https://eapi.binance.com",
			Description: "Binance Options API (Production)",
		},
	}

	// Override with custom server if provided
	if serverURL := os.Getenv("BINANCE_OPTIONS_REST_SERVER"); serverURL != "" {
		cfg.Servers[0].URL = serverURL
	}

	client := openapi.NewAPIClient(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get index price (current spot price)
	resp, _, err := client.OptionsAPI.GetIndexV1(ctx).Underlying(underlying).Execute()
	if err != nil {
		return 0, fmt.Errorf("failed to get index price for %s: %w", underlying, err)
	}

	if resp.IndexPrice == nil {
		return 0, fmt.Errorf("no index price returned for %s", underlying)
	}

	price, err := strconv.ParseFloat(*resp.IndexPrice, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse index price %s: %w", *resp.IndexPrice, err)
	}

	return price, nil
}

// abs returns the absolute value of x
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// clearSymbolCache clears the symbol cache (useful for testing)
func clearSymbolCache() {
	symbolCache.symbols = nil
	symbolCache.lastUpdate = time.Time{}
}