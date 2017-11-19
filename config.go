package monzo

import (
	"github.com/dghubble/oauth1"
)

// Config holds all config data for app
type Config struct {
	Monzo     MonzoConfig
	Splitwise SplitwiseConfig
}

// MonzoConfig holds config for Monzo's API
type MonzoConfig MonzoClient

// SplitwiseConfig holds config for Splitwise's API
type SplitwiseConfig struct {
	OAuthConfig oauth1.Config
	Token       oauth1.Token
}

func GetDefaultConfig() Config {
	config := Config{}
	return config
}
