package splitwise

import "github.com/dghubble/oauth1"

// SplitwiseConfig holds config for Splitwise's API
type SplitwiseConfig struct {
	OAuthConfig oauth1.Config
	Token       oauth1.Token
}
