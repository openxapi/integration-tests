module github.com/openxapi/integration-tests

go 1.24.1

replace github.com/openxapi/binance-go/ws => ../../../../../../binance-go/ws

replace github.com/openxapi/binance-go/rest => ../../../../../../binance-go/rest

require (
	github.com/openxapi/binance-go/rest v0.0.0-00010101000000-000000000000
	github.com/openxapi/binance-go/ws v0.0.0-00010101000000-000000000000
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
)
