package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	ms "github.com/cheahjs/monzosplitwise"
	"github.com/cheahjs/monzosplitwise/monzo"
	"github.com/cheahjs/monzosplitwise/splitwise"
)

func main() {
	fmt.Println("Starting MonzoSplitwise.")

	// Config loading
	fmt.Println("Loading config.json.")
	config, err := readConfig()
	if err != nil {
		fmt.Println("Failed to load config.")
		fmt.Println(err)
		return
	}
	// Getting Splitwise OAuth tokens
	if config.Splitwise.Token.Token == "" {
		tokens, err := splitwise.GetSplitwiseTokens(config.Splitwise.OAuthConfig)
		checkError(err)
		config.Splitwise.Token = *tokens
		saveConfig(config)
	}
	// Getting Monzo OAuth tokens
	if config.Monzo.AccessToken == "" {
		fmt.Println("Please sign in to Monzo at: ", monzo.GetMonzoAuthURL(config.Monzo.ClientID, "http://localhost/"))
		fmt.Printf("Choose whether to grant the application access.\nPaste " +
			"the code parameter from the address bar: ")
		fmt.Println()
		// Code is too long, need to split into 2
		var code string
		_, err = fmt.Scanf("%s\n", &code)
		checkError(err)
		var code2 string
		_, err = fmt.Scanf("%s\n", &code2)
		checkError(err)
		client, err := monzo.ExchangeAuth(config.Monzo.ClientID, config.Monzo.ClientSecret, "http://localhost/", code+code2)
		checkError(err)
		config.Monzo = monzo.MonzoConfig(*client)
		saveConfig(config)
	}
	runJob(config)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func runJob(config ms.Config) {
	monzoClient := monzo.MonzoClient(config.Monzo)
	// Refresh token if expired
	if !monzoClient.Authenticated() {
		err := monzoClient.RefreshAccessToken()
		checkError(err)
		config.Monzo = monzo.MonzoConfig(monzoClient)
		err = saveConfig(config)
		checkError(err)
	}

	// Get account to use, prefer CA over PP
	accounts, err := monzoClient.Accounts()
	checkError(err)
	account := accounts[0]
	for _, v := range accounts {
		if v.Type == "uk_retail" {
			account = v
		}
	}
	// Get transactions, 100 transactions and 15 days should be long enough
	// to get all transactions within context
	// This will break if we've made more than 100 transactions in 15 days.
	// TODO: Support pagination
	dateSince := time.Now().Add(time.Duration(-15*24) * time.Hour).Format(time.RFC3339)
	transactions, err := monzoClient.Transactions(
		account.ID,
		dateSince,
		"", 100)
	checkError(err)
	fmt.Printf("Fetched %v transactions\n", len(transactions))
	if len(transactions) >= 100 {
		fmt.Println("100 transactions fetched, might be missing some newer transactions!")
	}

	// Get current Splitwise user
	curUser, err := splitwise.GetCurrentUser(config.Splitwise)
	checkError(err)
	fmt.Println("Logged in as Splitwise user", curUser.Email)

	// Get Splitwise expenses
	expenses, err := splitwise.GetExpenses(config.Splitwise, "", dateSince, 100)
	checkError(err)
	fmt.Printf("Fetched %v expenses\n", len(expenses))

	// Get Splitwise groups
	groups, err := splitwise.GetGroups(config.Splitwise)
	checkError(err)
	fmt.Printf("Fetched %v groups\n", len(groups))

	// Find transactions with #splitwise as note
	tagged := getTaggedTransactions(transactions)

	for _, v := range tagged {
		tag := v.Tag
		tnx := v.Transaction

		// Check if expense already exists
		exists := false
		for _, exp := range expenses {
			if strings.Contains(exp.Details, tnx.ID) {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		var groupID string
		var groupName string
		var groupUsers []string

		// Get group ID
		switch tag {
		case "#splitwise":
		case "#splitwise-":
			groupID = "0"
			groupName = "Non-group expenses"
			groupUsers = append(groupUsers, fmt.Sprintf("%v", curUser.ID))
		default:
			groupName = strings.SplitN(tag, "-", 2)[1]
			group, err := findGroupByName(groups, groupName)
			if err != nil {
				fmt.Println("Group not found:", groupName)
				continue
			}
			groupID = fmt.Sprintf("%v", group.ID)
			for _, member := range group.Members {
				groupUsers = append(groupUsers, fmt.Sprintf("%v", member.ID))
			}
		}
		fmt.Println("Adding expense to group", groupName)
		expense, err := splitwise.AddExpense(
			config.Splitwise, "false", tnx.Amount, tnx.Currency, tnx.Merchant.Name,
			groupID, fmt.Sprintf("MonzoTransaction:%v", tnx.ID), tnx.Created,
			"split", fmt.Sprintf("%v", curUser.ID), groupUsers)
		checkError(err)
		fmt.Println("Added expense:")
		fmt.Println(expense)
	}

	fmt.Println("Done")
}

func findGroupByName(groups []splitwise.Group, name string) (*splitwise.Group, error) {
	normName := strings.ToLower(name)
	for _, v := range groups {
		groupName := strings.ToLower(strings.Replace(v.Name, " ", "", -1))
		if groupName == normName {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("No group found")
}

type taggedTransaction struct {
	Transaction monzo.Transaction
	Tag         string
}

func getTaggedTransactions(transactions []monzo.Transaction) []taggedTransaction {
	var tagged []taggedTransaction
	for _, v := range transactions {
		if v.Amount > 0 {
			// ignore credit transactions
			continue
		}
		notes := v.Notes
		fields := strings.Fields(notes)

		for _, field := range fields {
			if strings.Contains(field, "#splitwise") {
				tagged = append(tagged, taggedTransaction{v, field})
				break
			}
		}
	}
	return tagged
}

func readConfig() (ms.Config, error) {
	config := ms.Config{}
	// config.json exists
	if _, fileerr := os.Stat("config.json"); !os.IsNotExist(fileerr) {
		file, err := os.Open("config.json")
		if err != nil {
			return config, err
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			return config, err
		}
		return config, err
	}
	// config.json does not exist, create and return error
	err := saveConfig(ms.GetDefaultConfig())
	if err != nil {
		return config, err
	}
	return config, fmt.Errorf("config.json didn't exist, created")
}

func saveConfig(config ms.Config) error {
	file, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		return err
	}
	return nil
}
