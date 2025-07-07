// This file contains stubs for event types that might be referenced in the tests
// Since umfutures uses a generic event handler approach rather than specific 
// event handler methods, these stubs are not currently required.
// They are kept here for potential future use if the API evolves.

//go:build !skip_stubs

package wstest

// Note: The umfutures WebSocket implementation uses a generic event handler
// registration system rather than specific handler methods like:
// - client.HandleExecutionReport()
// - client.HandleBalanceUpdate()
// etc.
//
// Instead, it uses:
// - client.eventHandler.RegisterHandler(eventType string, handler func(interface{}) error)
//
// Therefore, no event stub types are required for compilation.