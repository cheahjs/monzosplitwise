package main

import (
	"encoding/json"
	"fmt"
	"math"
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

	// Find transactions with #splitwise-<group>
	for _, tnx := range transactions {
		notes := tnx.Notes
		parts := strings.Fields(notes)
		for _, part := range parts {
			if strings.HasPrefix(part, "#splitwise-") {
				// Check if we already have an expense with this transaction
				alreadyExists := false
				for _, expense := range expenses {
					if strings.Contains(expense.Details, tnx.ID) {
						alreadyExists = true
						break
					}
				}
				if alreadyExists {
					break
				}
				// Find corresponding group using name
				groupName := strings.ToLower(strings.Split(part, "-")[1])
				fmt.Println("Adding expense to group", groupName)
				for _, group := range groups {
					found := false
					if strings.ToLower(strings.Replace(group.Name, " ", "", -1)) == groupName {
						// Add expense to group
						expense, err := splitwise.AddExpense(
							config.Splitwise,
							"false",
							fmt.Sprintf("%v", (math.Abs(float64(tnx.Amount))/100.0)),
							tnx.Currency,
							tnx.Merchant.Name,
							fmt.Sprintf("%v", group.ID),
							fmt.Sprintf("MonzoTransaction:%v", tnx.ID),
							tnx.Created,
							"quickadd",
							fmt.Sprintf("%v", curUser.ID))
						checkError(err)
						fmt.Println("Added expense:")
						fmt.Println(expense)
						found = true
						break
					}
					if !found {
						fmt.Println("Warning, did not find any groups corresponding to", groupName)
					}
				}
				break
			}
		}
	}

	fmt.Println("Done")
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
