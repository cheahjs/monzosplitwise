package splitwise

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/oauth1"
	"github.com/rhymond/go-money"
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
	payment string, cost int, currencyCode, description, groupID, details, date,
	creationMethod, self string, users []string) (*Expense, error) {
	type expensesResponse struct {
		Expenses []Expense `json:"expenses"`
	}
	ctx := context.Background()
	httpClient := config.OAuthConfig.Client(ctx, &config.Token)

	stringFullCost := fmt.Sprintf("%v", (math.Abs(float64(cost)) / 100.0))
	costMoney := money.New(int64(cost), "GBP").Absolute()
	userCount := len(users)
	splits, err := costMoney.Split(userCount)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("payment", payment)
	form.Set("cost", stringFullCost)
	form.Set("currency_code", currencyCode)
	form.Set("description", description)
	form.Set("group_id", groupID)
	form.Set("details", details)
	form.Set("date", date)
	form.Set("creation_method", creationMethod)

	for i, user := range users {
		form.Set(fmt.Sprintf("users__%v__user_id", i), user)
		if user == self {
			form.Set(fmt.Sprintf("users__%v__paid_share", i), stringFullCost)
		} else {
			form.Set(fmt.Sprintf("users__%v__paid_share", i), "0")
		}
		form.Set(fmt.Sprintf("users__%v__owed_share", i), fmt.Sprintf("%v", (math.Abs(float64(splits[i].Amount()))/100.0)))
	}
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
