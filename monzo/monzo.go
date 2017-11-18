// Code adapted from https://github.com/sjwhitworth/gomondo

// Package monzo provides a (limited) Go interface for interacting with the Monzo API.
package monzo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// The root URL we will base all queries off of. Currently only production is supported.
	baseMonzoURL = "https://api.monzo.com"
	// OAuth response type
	responseType = "code"
	// OAuth grant types
	grantTypeRefresh  = "refresh_token"
	grantTypeAuthCode = "authorization_code"
)

var (
	// ErrUnauthenticatedRequest 401 response code
	ErrUnauthenticatedRequest = fmt.Errorf("your request was not sent with a valid token")
	// ErrNoTransactionFound No transaction found
	ErrNoTransactionFound = fmt.Errorf("no transaction found with ID")
)

// MonzoClient stores authentication data required for API calls
type MonzoClient struct {
	accessToken   string
	refreshToken  string
	authenticated bool
	clientID      string
	clientSecret  string
	expiryTime    time.Time
}

// GetMonzoAuthURL returns an OAuth URL that redirects to redirectURI
func GetMonzoAuthURL(clientID, redirectURI string) string {
	return fmt.Sprintf("https://auth.getmondo.co.uk/?client_id=%s&redirect_uri=%s&response_type=%s", clientID, redirectURI, responseType)
}

// ExchangeAuth exchanges the OAuth code for an access token and refresh token
func ExchangeAuth(clientID, clientSecret, redirectURI, code string) (*MonzoClient, error) {
	if clientID == "" || clientSecret == "" || redirectURI == "" || code == "" {
		return nil, fmt.Errorf("zero value passed to ExchangeAuth")
	}

	values := url.Values{}
	values.Set("grant_type", grantTypeAuthCode)
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("redirect_uri", redirectURI)
	values.Set("code", code)

	resp, err := http.PostForm(buildURL("oauth2/token"), values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, ErrUnauthenticatedRequest
	}

	response := tokenResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, fmt.Errorf(response.Error)
	}

	if response.ExpiresIn == 0 || response.TokenType == "" || response.AccessToken == "" {
		return nil, fmt.Errorf("failed to scan response correctly")
	}

	return &MonzoClient{
		authenticated: true,
		accessToken:   response.AccessToken,
		refreshToken:  response.RefreshToken,
		expiryTime:    time.Now().Add(time.Duration(response.ExpiresIn) * time.Second),
		clientID:      clientID,
		clientSecret:  clientSecret,
	}, nil
}

// RefreshToken refreshes the access token using the refresh token
func (m *MonzoClient) RefreshToken() error {
	values := url.Values{}
	values.Set("grant_type", grantTypeRefresh)
	values.Set("client_id", m.clientID)
	values.Set("client_secret", m.clientSecret)
	values.Set("refresh_token", m.refreshToken)

	resp, err := http.PostForm(buildURL("oauth2/token"), values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return ErrUnauthenticatedRequest
	}

	response := tokenResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &response); err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf(response.Error)
	}

	if response.ExpiresIn == 0 || response.TokenType == "" || response.AccessToken == "" {
		return fmt.Errorf("failed to scan response correctly")
	}

	m.authenticated = true
	m.accessToken = response.AccessToken
	m.refreshToken = response.RefreshToken
	m.expiryTime = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)

	return nil
}

// ExpiresAt returns the time that the current OAuth token expires and will have to be refreshed.
func (m *MonzoClient) ExpiresAt() time.Time {
	return m.expiryTime
}

// Authenticated returns true if there exists a valid token
func (m *MonzoClient) Authenticated() bool {
	if time.Now().Before(m.ExpiresAt()) {
		return true
	}
	m.authenticated = false
	return m.authenticated
}

// callWithAuth makes authenticated calls to the Monzo API.
func (m *MonzoClient) callWithAuth(methodType, URL string, params map[string]string) (*http.Response, error) {
	var resp *http.Response
	var err error

	// TODO: This is so hacky, clean up
	switch methodType {
	case "GET":
		req, err := http.NewRequest(methodType, buildURL(URL), nil)
		if err != nil {
			return nil, err
		}

		// If we have any parameters, add them here.
		if len(params) > 0 {
			query := req.URL.Query()
			for k, v := range params {
				query.Add(k, v)
			}
			req.URL.RawQuery = query.Encode()
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", m.accessToken))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			m.authenticated = false
			return nil, ErrUnauthenticatedRequest
		}

	case "POST":
		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}

		req, err := http.NewRequest(methodType, buildURL(URL), strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", m.accessToken))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			m.authenticated = false
			return nil, ErrUnauthenticatedRequest
		}
	}

	return resp, err
}

// Transactions returns a slice of Transactions, with the merchant expanded within the Transaction.
// This endpoint supports pagination. To paginate, provide the last Transacation.ID to the since parameter of the function, if the length of the results that are returned is equal to your limit.
func (m *MonzoClient) Transactions(accountID, since, before string, limit int) ([]Transaction, error) {
	type transactionsResponse struct {
		Transactions []Transaction `json:"transactions"`
	}

	params := map[string]string{
		"account_id": accountID,
		"expand[]":   "merchant",
		"limit":      fmt.Sprintf("%v", limit),
		"since":      since,
		"before":     before,
	}

	resp, err := m.callWithAuth("GET", "transactions", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := transactionsResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return response.Transactions, nil
}

// TransactionByID obtains a Monzo Transaction by a specific transaction ID.
func (m *MonzoClient) TransactionByID(accountID, transactionID string) (*Transaction, error) {
	type transactionByIDResponse struct {
		Transaction Transaction `json:"transaction"`
	}

	params := map[string]string{
		"account_id": accountID,
		"expand[]":   "merchant",
	}

	resp, err := m.callWithAuth("GET", fmt.Sprintf("transactions/%s", transactionID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, ErrNoTransactionFound
	}

	response := transactionByIDResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return &response.Transaction, nil
}

// Accounts returns a list of accounts
func (m *MonzoClient) Accounts() ([]Account, error) {
	type accountsResponse struct {
		Accounts []Account `json:"accounts"`
	}

	resp, err := m.callWithAuth("GET", "accounts", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	acresp := accountsResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &acresp); err != nil {
		return nil, err
	}

	return acresp.Accounts, nil
}

func buildURL(path string) string {
	return fmt.Sprintf("%v/%v", baseMonzoURL, path)
}
