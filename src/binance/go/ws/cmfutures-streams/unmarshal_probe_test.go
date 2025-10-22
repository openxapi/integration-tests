package streamstest

import (
    "encoding/json"
    "os"
    "testing"
    sdkmodels "github.com/openxapi/binance-go/ws/cmfutures-streams/models"
)

// Minimal probe to validate whether the SDK model can decode a real indexPrice_kline payload
func Test_Unmarshal_IndexPriceKline_Sample(t *testing.T) {
    if os.Getenv("RUN_UNMARSHAL_PROBE") != "1" {
        t.Skip("skipping probe; set RUN_UNMARSHAL_PROBE=1 to run")
    }
    sample := []byte(`{
        "e": "indexPrice_kline",
        "E": 1761068511005,
        "ps": "BTCUSD",
        "k": {"t":1761068460000,"T":1761068519999,"s":"0","i":"1m","f":1761068460000,"L":1761068511001,
              "o":"112540.16287576","c":"112538.09190858","h":"112596.94362126","l":"112511.53875234",
              "v":"0","n":52,"x":false,"q":"0","V":"0","Q":"0","B":"0"}
    }`)

    // sanity: it is valid JSON object
    var generic map[string]any
    if err := json.Unmarshal(sample, &generic); err != nil {
        t.Fatalf("generic unmarshal failed: %v", err)
    }

    // try to unmarshal into SDK model
    var ev sdkmodels.IndexKlineEvent
    if err := json.Unmarshal(sample, &ev); err != nil {
        t.Fatalf("SDK IndexKlineEvent unmarshal failed: %v", err)
    }
    if ev.EventType != "indexPrice_kline" || ev.Pair == "" || ev.KlineData.Interval == "" {
        t.Fatalf("unexpected decoded fields: %+v", ev)
    }
}
