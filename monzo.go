// Code adapted from https://github.com/sjwhitworth/gomondo

// Package go-monzo provides a Go interface for interacting with the Monzo API.
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

var (
    // The root URL we will base all queries off of. Currently only production is supported.
    BaseMonzoURL = "https://api.monzo.com"

    // OAuth response type
    ResponseType = "code"

    // OAuth grant type.
    GrantTypeRefresh = "refresh_token"

    // 401 response code
    ErrUnauthenticatedRequest = fmt.Errorf("your request was not sent with a valid token")

    // No transaction found
    ErrNoTransactionFound = fmt.Errorf("no transaction found with ID")
)

type MonzoClient struct {
    accessToken   string
    refreshToken  string
    authenticated bool
    expiryTime    time.Time
}


func GetMonzoAuthUrl(clientId, redirectUri string) string {
    return fmt.Sprintf("https://auth.getmondo.co.uk/?client_id=%s&redirect_uri=%s&response_type=%s", clientId, redirectUri, ResponseType)
}

// Function Authenticate authenticates the user using the oath flow, returning an authenticated MonzoClient
func Authenticate(clientId, clientSecret, redirect_uri, password string) (*MonzoClient, error) {
    if clientId == "" || clientSecret == "" || username == "" || password == "" {
        return nil, fmt.Errorf("zero value passed to Authenticate")
    }

    values := url.Values{}
    values.Set("grant_type", GrantTypeRefresh)
    values.Set("client_id", clientId)
    values.Set("client_secret", clientSecret)
    values.Set("username", username)
    values.Set("password", password)

    resp, err := http.PostForm(buildUrl("oauth2/token"), values)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        return nil, ErrUnauthenticatedRequest
    }

    tresp := tokenResponse{}
    b, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal(b, &tresp); err != nil {
        return nil, err
    }

    if tresp.Error != "" {
        return nil, fmt.Errorf(tresp.Error)
    }

    if tresp.ExpiresIn == 0 || tresp.TokenType == "" || tresp.AccessToken == "" {
        return nil, fmt.Errorf("failed to scan response correctly")
    }

    return &MonzoClient{
        authenticated: true,
        accessToken:   tresp.AccessToken,
        expiryTime:    time.Now().Add(time.Duration(tresp.ExpiresIn) * time.Second),
    }, nil
}

// ExpiresAt returns the time that the current oauth token expires and will have to be refreshed.
func (m *MonzoClient) ExpiresAt() time.Time {
    return m.expiryTime
}
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
        req, err := http.NewRequest(methodType, buildUrl(URL), nil)
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

        req, err := http.NewRequest(methodType, buildUrl(URL), strings.NewReader(form.Encode()))
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

// Transactions returns a slice of Transactions, with the merchant expanded within the Transaction. This endpoint supports pagination. To paginate, provide the last Transacation.ID to the since parameter of the function, if the length of the results that are returned is equal to your limit.
func (m *MonzoClient) Transactions(accountId, since, before string, limit int) ([]Transaction, error) {
    type transactionsResponse struct {
        Transactions []Transaction `json:"transactions"`
    }

    params := map[string]string{
        "account_id": accountId,
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

    tresp := transactionsResponse{}
    b, err := ioutil.ReadAll(resp.Body)
    if err := json.Unmarshal(b, &tresp); err != nil {
        return nil, err
    }

    return tresp.Transactions, nil
}

// TransactionByID obtains a Monzo Transaction by a specific transaction ID.
func (m *MonzoClient) TransactionByID(accountId, transactionId string) (*Transaction, error) {
    type transactionByIDResponse struct {
        Transaction Transaction `json:"transaction"`
    }

    params := map[string]string{
        "account_id": accountId,
        "expand[]":   "merchant",
    }

    resp, err := m.callWithAuth("GET", fmt.Sprintf("transactions/%s", transactionId), params)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404 {
        return nil, ErrNoTransactionFound
    }

    tresp := transactionByIDResponse{}
    b, err := ioutil.ReadAll(resp.Body)
    if err := json.Unmarshal(b, &tresp); err != nil {
        return nil, err
    }

    return &tresp.Transaction, nil
}

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

func buildUrl(path string) string {
    return fmt.Sprintf("%v/%v", BaseMonzoURL, path)
}