package wstest

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	spot "github.com/openxapi/binance-go/ws/spot"
)

// signAndApply signs the provided params map and writes apiKey/signature/timestamp
// into the target params struct (must be a pointer to the request Params field).
func signAndApply(t testing.TB, signer *spot.RequestSigner, authType spot.AuthType, params map[string]interface{}, target interface{}) map[string]interface{} {
	t.Helper()
	if authType == spot.AuthTypeNone {
		return params
	}
	if signer == nil {
		t.Fatalf("signer is required for auth type %s", authType)
	}
	if params == nil {
		params = make(map[string]interface{})
	}
	if err := signer.SignRequest(params, authType); err != nil {
		t.Fatalf("sign request (%s) failed: %v", authType, err)
	}
	if target != nil {
		applyAuthFields(t, target, params)
	}
	return params
}

func applyAuthFields(t testing.TB, target interface{}, params map[string]interface{}) {
	t.Helper()
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		t.Fatalf("target must be non-nil pointer; got %T", target)
	}
	val = val.Elem()
	if !val.IsValid() {
		t.Fatalf("invalid target value %T", target)
	}

	setStringField(val, "ApiKey", params["apiKey"])
	setStringField(val, "Signature", params["signature"])
	setInt64Field(val, "Timestamp", params["timestamp"])
}

func setStringField(structVal reflect.Value, field string, value interface{}) {
	f := structVal.FieldByName(field)
	if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.String {
		return
	}
	if value == nil {
		return
	}
	if s, ok := value.(string); ok {
		f.SetString(s)
	}
}

func setInt64Field(structVal reflect.Value, field string, value interface{}) {
	f := structVal.FieldByName(field)
	if !f.IsValid() || !f.CanSet() {
		return
	}
	switch f.Kind() {
	case reflect.Int64, reflect.Int, reflect.Int32:
	default:
		return
	}
	if value == nil {
		return
	}
	switch v := value.(type) {
	case int64:
		f.SetInt(v)
	case int:
		f.SetInt(int64(v))
	case int32:
		f.SetInt(int64(v))
	case uint64:
		f.SetInt(int64(v))
	case uint:
		f.SetInt(int64(v))
	case fmt.Stringer:
		var parsed int64
		if _, err := fmt.Sscan(v.String(), &parsed); err == nil {
			f.SetInt(parsed)
		}
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.SetInt(parsed)
		}
	}
}
