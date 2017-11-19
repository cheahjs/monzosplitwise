package main

import (
	"encoding/json"
	"fmt"
	"os"

	monzo "github.com/cheahjs/monzosplitwise"
)

func main() {
	fmt.Println("Starting monzosplitwise.")

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
		tokens, err := monzo.GetSplitwiseTokens(config.Splitwise.OAuthConfig)
		if err != nil {
			fmt.Println(err)
			return
		}
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
		if err != nil {
			fmt.Println(err)
			return
		}
		var code2 string
		_, err = fmt.Scanf("%s\n", &code2)
		if err != nil {
			fmt.Println(err)
			return
		}
		client, err := monzo.ExchangeAuth(config.Monzo.ClientID, config.Monzo.ClientSecret, "http://localhost/", code+code2)
		if err != nil {
			fmt.Println(err)
			return
		}
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

func runJob(config monzo.Config) {
	monzoClient := monzo.MonzoClient(config.Monzo)
	// Refresh token if expired
	if !monzoClient.Authenticated() {
		err := monzoClient.RefreshAccessToken()
		checkError(err)
		config.Monzo = monzo.MonzoConfig(monzoClient)
		err = saveConfig(config)
		checkError(err)
	}
	accounts, err := monzoClient.Accounts()
	checkError(err)
	fmt.Println(accounts)
}

func readConfig() (monzo.Config, error) {
	config := monzo.Config{}
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
	err := saveConfig(monzo.GetDefaultConfig())
	if err != nil {
		return config, err
	}
	return config, fmt.Errorf("config.json didn't exist, created")
}

func saveConfig(config monzo.Config) error {
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
