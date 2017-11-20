package monzosplitwise

import (
	"github.com/cheahjs/monzosplitwise/monzo"
	"github.com/cheahjs/monzosplitwise/splitwise"
)

// Config holds all config data for app
type Config struct {
	Monzo     monzo.MonzoConfig
	Splitwise splitwise.SplitwiseConfig
}

// GetDefaultConfig returns a default config object with blank fields
func GetDefaultConfig() Config {
	config := Config{}
	return config
}
