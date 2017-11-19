package monzo

import (
	"fmt"

	"github.com/dghubble/oauth1"
)

// GetSplitwiseTokens interactively requests for an OAuth code and returns an access token
func GetSplitwiseTokens(config oauth1.Config) (*oauth1.Token, error) {
	// ctx := context.Background()
	requestToken, requestSecret, err := config.RequestToken()
	if err != nil {
		return nil, err
	}
	authorizationURL, err := config.AuthorizationURL(requestToken)
	if err != nil {
		return nil, err
	}
	fmt.Println("Please sign in at: ", authorizationURL.String())
	fmt.Printf("Choose whether to grant the application access.\nPaste " +
		"the oauth_verifier parameter from the " +
		"address bar: ")
	var verifier string
	_, err = fmt.Scanf("%s", &verifier)
	accessToken, accessSecret, err := config.AccessToken(requestToken, requestSecret, verifier)
	if err != nil {
		return nil, err
	}
	return oauth1.NewToken(accessToken, accessSecret), err
}
