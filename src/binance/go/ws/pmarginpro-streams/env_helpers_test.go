package streamstest

import (
	"os"
	"strings"
)

func apiKeyFromEnv() string {
	return strings.TrimSpace(os.Getenv("BINANCE_API_KEY"))
}

func secretKeyFromEnv() string {
	return strings.TrimSpace(os.Getenv("BINANCE_SECRET_KEY"))
}

func preferredListenKeyFromEnv() string {
	return strings.TrimSpace(os.Getenv("BINANCE_LISTEN_KEY"))
}
