module github.com/openxapi/integration-tests

go 1.24.1

replace github.com/openxapi/binance-go/ws => ../../../../../../binance-go/ws

require (
	github.com/openxapi/binance-go/ws v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
