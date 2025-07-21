package main

import (
	"testing"
	"time"
)

// testAccountInfo tests the options account info endpoint
func testAccountInfo(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetAccountV1(ctx).
		Timestamp(time.Now().UnixMilli()).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetAccountV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetAccountV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Options Account Info:")
	if resp.RiskLevel != nil {
		t.Logf("  Risk Level: %s", *resp.RiskLevel)
	}
	if resp.Time != nil {
		t.Logf("  Time: %d (%s)", *resp.Time, time.UnixMilli(*resp.Time))
	}
	
	// Log asset information
	if len(resp.Asset) > 0 {
		t.Logf("  Assets:")
		for i, asset := range resp.Asset {
			t.Logf("    Asset %d:", i+1)
			if asset.Asset != nil {
				t.Logf("      Currency: %s", *asset.Asset)
			}
			if asset.Equity != nil {
				t.Logf("      Equity: %s", *asset.Equity)
			}
			if asset.Available != nil {
				t.Logf("      Available: %s", *asset.Available)
			}
			if asset.Locked != nil {
				t.Logf("      Locked: %s", *asset.Locked)
			}
			if asset.MarginBalance != nil {
				t.Logf("      Margin Balance: %s", *asset.MarginBalance)
			}
			if asset.UnrealizedPNL != nil {
				t.Logf("      Unrealized PNL: %s", *asset.UnrealizedPNL)
			}
		}
	}
	
	// Log Greek information
	if len(resp.Greek) > 0 {
		t.Logf("  Greeks:")
		for i, greek := range resp.Greek {
			t.Logf("    Greek %d:", i+1)
			if greek.Underlying != nil {
				t.Logf("      Underlying: %s", *greek.Underlying)
			}
			if greek.Delta != nil {
				t.Logf("      Delta: %s", *greek.Delta)
			}
			if greek.Gamma != nil {
				t.Logf("      Gamma: %s", *greek.Gamma)
			}
			if greek.Theta != nil {
				t.Logf("      Theta: %s", *greek.Theta)
			}
			if greek.Vega != nil {
				t.Logf("      Vega: %s", *greek.Vega)
			}
		}
	}
}

// testPositionInfo tests the options position info endpoint
func testPositionInfo(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetPositionV1(ctx).
		Timestamp(time.Now().UnixMilli()).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetPositionV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetPositionV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Options Positions: %d positions", len(resp))
	
	// Log first few positions
	if len(resp) > 0 {
		for i, position := range resp {
			if i >= 5 { // Only log first 5 positions
				break
			}
			t.Logf("Position %d:", i+1)
			if position.Symbol != nil {
				t.Logf("  Symbol: %s", *position.Symbol)
			}
			if position.Side != nil {
				t.Logf("  Side: %s", *position.Side)
			}
			if position.Quantity != nil {
				t.Logf("  Quantity: %s", *position.Quantity)
			}
			if position.EntryPrice != nil {
				t.Logf("  Entry Price: %s", *position.EntryPrice)
			}
			if position.MarkPrice != nil {
				t.Logf("  Mark Price: %s", *position.MarkPrice)
			}
			if position.UnrealizedPNL != nil {
				t.Logf("  Unrealized PNL: %s", *position.UnrealizedPNL)
			}
			if position.QuoteAsset != nil {
				t.Logf("  Quote Asset: %s", *position.QuoteAsset)
			}
			if position.PositionCost != nil {
				t.Logf("  Position Cost: %s", *position.PositionCost)
			}
			if position.ReducibleQty != nil {
				t.Logf("  Reducible Qty: %s", *position.ReducibleQty)
			}
			if position.StrikePrice != nil {
				t.Logf("  Strike Price: %s", *position.StrikePrice)
			}
			if position.ExpiryDate != nil {
				t.Logf("  Expiry Date: %d (%s)", *position.ExpiryDate, time.UnixMilli(*position.ExpiryDate))
			}
		}
	} else {
		t.Log("No positions found")
	}
}

// testMarginAccountInfo tests the options margin account info endpoint
func testMarginAccountInfo(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetMarginAccountV1(ctx).
		Timestamp(time.Now().UnixMilli()).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetMarginAccountV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetMarginAccountV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Options Margin Account Info:")
	if resp.Time != nil {
		t.Logf("  Time: %d (%s)", *resp.Time, time.UnixMilli(*resp.Time))
	}
	
	// Log asset information
	if len(resp.Asset) > 0 {
		t.Logf("  Assets:")
		for i, asset := range resp.Asset {
			t.Logf("    Asset %d:", i+1)
			if asset.Asset != nil {
				t.Logf("      Currency: %s", *asset.Asset)
			}
			if asset.Equity != nil {
				t.Logf("      Equity: %s", *asset.Equity)
			}
			if asset.Available != nil {
				t.Logf("      Available: %s", *asset.Available)
			}
			if asset.MarginBalance != nil {
				t.Logf("      Margin Balance: %s", *asset.MarginBalance)
			}
			if asset.UnrealizedPNL != nil {
				t.Logf("      Unrealized PNL: %s", *asset.UnrealizedPNL)
			}
			if asset.InitialMargin != nil {
				t.Logf("      Initial Margin: %s", *asset.InitialMargin)
			}
			if asset.MaintMargin != nil {
				t.Logf("      Maintenance Margin: %s", *asset.MaintMargin)
			}
			if asset.LpProfit != nil {
				t.Logf("      LP Profit: %s", *asset.LpProfit)
			}
		}
	}
	
	// Log Greek information
	if len(resp.Greek) > 0 {
		t.Logf("  Greeks:")
		for i, greek := range resp.Greek {
			t.Logf("    Greek %d:", i+1)
			if greek.Underlying != nil {
				t.Logf("      Underlying: %s", *greek.Underlying)
			}
			if greek.Delta != nil {
				t.Logf("      Delta: %s", *greek.Delta)
			}
			if greek.Gamma != nil {
				t.Logf("      Gamma: %s", *greek.Gamma)
			}
			if greek.Theta != nil {
				t.Logf("      Theta: %s", *greek.Theta)
			}
			if greek.Vega != nil {
				t.Logf("      Vega: %s", *greek.Vega)
			}
		}
	}
}

// testAccountBill tests the account funding flow endpoint
func testAccountBill(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetBillV1(ctx).
		Currency("USDT").
		Timestamp(time.Now().UnixMilli()).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetBillV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetBillV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Account Bill: %d entries", len(resp))
	
	// Log first few bill entries
	if len(resp) > 0 {
		for i, bill := range resp {
			if i >= 5 { // Only log first 5 entries
				break
			}
			t.Logf("Bill %d:", i+1)
			if bill.Id != nil {
				t.Logf("  ID: %d", *bill.Id)
			}
			if bill.Asset != nil {
				t.Logf("  Asset: %s", *bill.Asset)
			}
			if bill.Type != nil {
				t.Logf("  Type: %s", *bill.Type)
			}
			if bill.Amount != nil {
				t.Logf("  Amount: %s", *bill.Amount)
			}
			if bill.CreateDate != nil {
				t.Logf("  Create Date: %d (%s)", *bill.CreateDate, time.UnixMilli(*bill.CreateDate))
			}
		}
	} else {
		t.Log("No bill entries found")
	}
}

// testUserTrades tests the user trades endpoint
func testUserTrades(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetUserTradesV1(ctx).
		Timestamp(time.Now().UnixMilli()).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetUserTradesV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetUserTradesV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("User Trades: %d trades", len(resp))
	
	// Log first few trades
	if len(resp) > 0 {
		for i, trade := range resp {
			if i >= 5 { // Only log first 5 trades
				break
			}
			t.Logf("Trade %d:", i+1)
			if trade.Id != nil {
				t.Logf("  ID: %d", *trade.Id)
			}
			if trade.Symbol != nil {
				t.Logf("  Symbol: %s", *trade.Symbol)
			}
			if trade.Price != nil {
				t.Logf("  Price: %s", *trade.Price)
			}
			if trade.Quantity != nil {
				t.Logf("  Quantity: %s", *trade.Quantity)
			}
			if trade.Fee != nil {
				t.Logf("  Fee: %s", *trade.Fee)
			}
			if trade.RealizedProfit != nil {
				t.Logf("  Realized Profit: %s", *trade.RealizedProfit)
			}
			if trade.Side != nil {
				t.Logf("  Side: %s", *trade.Side)
			}
			if trade.Time != nil {
				t.Logf("  Time: %d (%s)", *trade.Time, time.UnixMilli(*trade.Time))
			}
		}
	} else {
		t.Log("No trades found")
	}
}

// testBlockUserTrades tests the block user trades endpoint
func testBlockUserTrades(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetBlockUserTradesV1(ctx).
		Timestamp(time.Now().UnixMilli()).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetBlockUserTradesV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetBlockUserTradesV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Block User Trades: %d trades", len(resp))
	
	// Log first few block trades
	if len(resp) > 0 {
		for i, trade := range resp {
			if i >= 5 { // Only log first 5 trades
				break
			}
			t.Logf("Block Trade %d:", i+1)
			if trade.BlockTradeSettlementKey != nil {
				t.Logf("  Settlement Key: %s", *trade.BlockTradeSettlementKey)
			}
			if trade.CrossType != nil {
				t.Logf("  Cross Type: %s", *trade.CrossType)
			}
			if trade.ParentOrderId != nil {
				t.Logf("  Parent Order ID: %s", *trade.ParentOrderId)
			}
			
			// Log legs information
			if len(trade.Legs) > 0 {
				t.Logf("  Legs (%d):", len(trade.Legs))
				for j, leg := range trade.Legs {
					if j >= 3 { // Only log first 3 legs
						break
					}
					t.Logf("    Leg %d:", j+1)
					if leg.Symbol != nil {
						t.Logf("      Symbol: %s", *leg.Symbol)
					}
					if leg.TradePrice != nil {
						t.Logf("      Trade Price: %f", *leg.TradePrice)
					}
					if leg.TradeQty != nil {
						t.Logf("      Trade Qty: %f", *leg.TradeQty)
					}
					if leg.OrderSide != nil {
						t.Logf("      Order Side: %s", *leg.OrderSide)
					}
				}
			}
		}
	} else {
		t.Log("No block trades found")
	}
}

// testExerciseRecord tests the exercise record endpoint
func testExerciseRecord(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetExerciseRecordV1(ctx).
		Timestamp(time.Now().UnixMilli()).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetExerciseRecordV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetExerciseRecordV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Exercise Records: %d records", len(resp))
	
	// Log first few exercise records
	if len(resp) > 0 {
		for i, record := range resp {
			if i >= 5 { // Only log first 5 records
				break
			}
			t.Logf("Exercise Record %d:", i+1)
			if record.Id != nil {
				t.Logf("  ID: %s", *record.Id)
			}
			if record.Symbol != nil {
				t.Logf("  Symbol: %s", *record.Symbol)
			}
			if record.PositionSide != nil {
				t.Logf("  Position Side: %s", *record.PositionSide)
			}
			if record.Quantity != nil {
				t.Logf("  Quantity: %s", *record.Quantity)
			}
			if record.Amount != nil {
				t.Logf("  Amount: %s", *record.Amount)
			}
			if record.Fee != nil {
				t.Logf("  Fee: %s", *record.Fee)
			}
			if record.CreateDate != nil {
				t.Logf("  Create Date: %d (%s)", *record.CreateDate, time.UnixMilli(*record.CreateDate))
			}
			if record.PriceScale != nil {
				t.Logf("  Price Scale: %d", *record.PriceScale)
			}
			if record.QuantityScale != nil {
				t.Logf("  Quantity Scale: %d", *record.QuantityScale)
			}
			if record.OptionSide != nil {
				t.Logf("  Option Side: %s", *record.OptionSide)
			}
			if record.ExercisePrice != nil {
				t.Logf("  Exercise Price: %s", *record.ExercisePrice)
			}
			if record.MarkPrice != nil {
				t.Logf("  Mark Price: %s", *record.MarkPrice)
			}
		}
	} else {
		t.Log("No exercise records found")
	}
}

// testIncomeAsync tests the income async endpoint
func testIncomeAsync(t *testing.T) {
	client, ctx := getTestClientAndContext(t)
	
	rateLimiter.WaitForRateLimit()
	
	resp, httpResp, err := client.OptionsAPI.GetIncomeAsynV1(ctx).
		Execute()
	
	if handleOptionsSpecificErrors(t, err, httpResp, "GetIncomeAsynV1") {
		return
	}
	
	if err != nil {
		t.Fatalf("GetIncomeAsynV1 failed: %v", err)
	}
	
	if httpResp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", httpResp.StatusCode)
	}
	
	if resp == nil {
		t.Fatal("Response should not be nil")
	}
	
	t.Logf("Income Async Response:")
	if resp.DownloadId != nil {
		t.Logf("  Download ID: %s", *resp.DownloadId)
	}
	if resp.AvgCostTimestampOfLast30d != nil {
		t.Logf("  Avg Cost Timestamp Of Last 30d: %d", *resp.AvgCostTimestampOfLast30d)
	}
}