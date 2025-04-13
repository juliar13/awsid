package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type AccountInfo struct {
	AliasName string `json:"alias_name"`
	AccountID string `json:"account_id"`
}

type AccountInfoList struct {
	Accounts []AccountInfo `json:"account_info"`
}

func main() {
	var jsonOutput bool
	var rootCmd = &cobra.Command{
		Use:   "awsid [alias_name]",
		Short: "Get AWS account ID from alias name",
		Long:  "A CLI tool to get AWS account ID from alias name",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// Get home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
				os.Exit(1)
			}

			// Path to account_info file
			accountInfoPath := homeDir + "/.aws/account_info"

			// Read account_info file
			accounts, err := readAccountInfo(accountInfoPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading account info: %v\n", err)
				os.Exit(1)
			}

			// If no arguments, list all accounts
			if len(args) == 0 {
				if jsonOutput {
					outputJSON(accounts)
				} else {
					for _, account := range accounts {
						fmt.Printf("%s: %s\n", account.AliasName, account.AccountID)
					}
				}
				return
			}

			// If argument is provided, search for matching accounts
			searchTerm := args[0]
			matchingAccounts := []AccountInfo{}

			for _, account := range accounts {
				if strings.HasPrefix(account.AliasName, searchTerm) {
					matchingAccounts = append(matchingAccounts, account)
				}
			}

			// If JSON output is requested
			if jsonOutput {
				outputJSON(matchingAccounts)
				return
			}

			// If exact match is found and not JSON, just print the account ID
			for _, account := range matchingAccounts {
				if account.AliasName == searchTerm {
					fmt.Println(account.AccountID)
					return
				}
			}

			// If partial matches found, print them
			if len(matchingAccounts) > 0 {
				for _, account := range matchingAccounts {
					fmt.Printf("%s: %s\n", account.AliasName, account.AccountID)
				}
				return
			}

			// No matches found
			fmt.Fprintf(os.Stderr, "No account found with alias name: %s\n", searchTerm)
			os.Exit(1)
		},
	}

	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func readAccountInfo(filePath string) ([]AccountInfo, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	accounts := []AccountInfo{}
	scanner := bufio.NewScanner(file)

	// Read line by line
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split line by space or tab
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			accounts = append(accounts, AccountInfo{
				AliasName: parts[0],
				AccountID: parts[1],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func outputJSON(accounts []AccountInfo) {
	output := AccountInfoList{
		Accounts: accounts,
	}

	jsonData, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}
