package monzo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
)

const (
	GetExpensesURL    = "https://secure.splitwise.com/api/v3.0/get_expenses"
	GetGroupsURL      = "https://secure.splitwise.com/api/v3.0/get_groups"
	CreateExpenseURL  = "https://secure.splitwise.com/api/v3.0/create_expense"
	GetCurrentUserURL = "https://secure.splitwise.com/api/v3.0/get_current_user"
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

type Expense struct {
	ID              int       `json:"id"`
	GroupID         int       `json:"group_id"`
	FriendshipID    int       `json:"friendship_id"`
	ExpenseBundleID int       `json:"expense_bundle_id"`
	Description     string    `json:"description"`
	Details         string    `json:"details"`
	Payment         bool      `json:"payment"`
	Cost            string    `json:"cost"`
	Date            time.Time `json:"date"`
	CreatedAt       time.Time `json:"created_at"`
	//CreatedBy        string    `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	//UpdatedBy        string    `json:"updated_by"`
	DeletedAt time.Time `json:"deleted_at"`
	//DeletedBy        string    `json:"deleted_by"`
}

func GetExpenses(config SplitwiseConfig, groupID, datedAfter string, limit int) ([]Expense, error) {
	type expensesResponse struct {
		Expenses []Expense `json:"expenses"`
	}

	params := map[string]string{
		"limit":       fmt.Sprintf("%v", limit),
		"dated_after": datedAfter,
		"group_id":    groupID,
	}

	ctx := context.Background()
	httpClient := config.OAuthConfig.Client(ctx, &config.Token)

	url, err := buildURLParams(GetExpensesURL, params)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := expensesResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return response.Expenses, nil
}

type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	GroupType string    `json:"group_type"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetGroups(config SplitwiseConfig) ([]Group, error) {
	type groupsResponse struct {
		Groups []Group `json:"groups"`
	}
	ctx := context.Background()
	httpClient := config.OAuthConfig.Client(ctx, &config.Token)

	resp, err := httpClient.Get(GetGroupsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := groupsResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return response.Groups, nil
}

func AddExpense(config SplitwiseConfig,
	payment, cost, currencyCode, description, groupID, details, date, creationMethod,
	userID string) (*Expense, error) {
	type expensesResponse struct {
		Expenses []Expense `json:"expenses"`
	}
	ctx := context.Background()
	httpClient := config.OAuthConfig.Client(ctx, &config.Token)

	form := url.Values{}
	form.Set("payment", payment)
	form.Set("cost", cost)
	form.Set("currency_code", currencyCode)
	form.Set("description", description)
	form.Set("group_id", groupID)
	form.Set("details", details)
	form.Set("date", date)
	form.Set("creation_method", creationMethod)
	form.Set("users__0__user_id", userID)
	form.Set("users__0__paid_share", cost)
	form.Set("users__0__owed_share", cost)
	fmt.Println(form)

	req, err := http.NewRequest("POST", CreateExpenseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	response := expensesResponse{}
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return &response.Expenses[0], nil
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	// We don't care about the other attributes for now
}

func GetCurrentUser(config SplitwiseConfig) (*User, error) {
	type userReponse struct {
		User User `json:"user"`
	}
	ctx := context.Background()
	httpClient := config.OAuthConfig.Client(ctx, &config.Token)

	resp, err := httpClient.Get(GetCurrentUserURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := userReponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return &response.User, nil
}

func buildURLParams(URL string, params map[string]string) (string, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	// If we have any parameters, add them here.
	if len(params) > 0 {
		query := req.URL.Query()
		for k, v := range params {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}
	return req.URL.String(), nil
}
