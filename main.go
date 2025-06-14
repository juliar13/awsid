package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type AccountInfo struct {
	AliasName string `json:"alias_name"`
	AccountID string `json:"account_id"`
}

type AccountInfoList struct {
	Accounts []AccountInfo `json:"account_info"`
}

const Version = "0.3.0"

func main() {
	var jsonOutput bool
	var tableOutput bool
	var csvOutput bool
	var rootCmd = &cobra.Command{
		Use:     "awsid [alias_name]",
		Short:   "Get AWS account ID from alias name",
		Long:    "A CLI tool to get AWS account ID from alias name",
		Version: Version,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// Get home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
				os.Exit(1)
			}

			// Path to account_info file
			accountInfoPath := filepath.Join(homeDir, ".aws", "account_info")

			// Try to update account info from AWS Organizations
			err = updateAccountInfoFromAWS(accountInfoPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to update account info from AWS: %v\n", err)
			}

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
				} else if tableOutput {
					outputTable(accounts)
				} else if csvOutput {
					outputCSV(accounts)
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

			// Check for exact match first
			exactMatch := []AccountInfo{}
			for _, account := range matchingAccounts {
				if account.AliasName == searchTerm {
					exactMatch = append(exactMatch, account)
					break
				}
			}

			// If exact match found
			if len(exactMatch) > 0 {
				if tableOutput {
					outputTable(exactMatch)
				} else if csvOutput {
					outputCSV(exactMatch)
				} else {
					fmt.Println(exactMatch[0].AccountID)
				}
				return
			}

			// If table output is requested for partial matches
			if tableOutput {
				outputTable(matchingAccounts)
				return
			}

			// If CSV output is requested for partial matches
			if csvOutput {
				outputCSV(matchingAccounts)
				return
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
	rootCmd.Flags().BoolVar(&tableOutput, "table", false, "Output in table format")
	rootCmd.Flags().BoolVar(&csvOutput, "csv", false, "Output in CSV format")

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

	// Read as CSV
	csvReader := csv.NewReader(file)
	csvReader.Comment = '#'
	csvReader.TrimLeadingSpace = true
	
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}
	
	// Process CSV records
	for i, record := range records {
		// Skip header row if it looks like a header
		if i == 0 && (record[0] == "alias_name" || record[0] == "AliasName") {
			continue
		}
		
		if len(record) >= 2 && record[0] != "" && record[1] != "" {
			accounts = append(accounts, AccountInfo{
				AliasName: strings.TrimSpace(record[0]),
				AccountID: strings.TrimSpace(record[1]),
			})
		}
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

func outputTable(accounts []AccountInfo) {
	table := tablewriter.NewTable(os.Stdout)
	table.Header("Alias Name", "Account ID")

	for _, account := range accounts {
		table.Append([]any{account.AliasName, account.AccountID})
	}

	table.Render()
}
func outputCSV(accounts []AccountInfo) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"alias_name", "account_id"})

	// Write data
	for _, account := range accounts {
		writer.Write([]string{account.AliasName, account.AccountID})
	}
}

func updateAccountInfoFromAWS(filePath string) error {
	// Create .aws directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Organizations client
	client := organizations.NewFromConfig(cfg)

	// List accounts
	ctx := context.TODO()
	result, err := client.ListAccounts(ctx, &organizations.ListAccountsInput{})
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	// Prepare account info
	var accounts []AccountInfo
	for _, account := range result.Accounts {
		if account.Id != nil && account.Name != nil {
			accounts = append(accounts, AccountInfo{
				AliasName: *account.Name,
				AccountID: *account.Id,
			})
		}
	}

	// Save to CSV file
	return saveAccountInfoToCSV(filePath, accounts)
}

func saveAccountInfoToCSV(filePath string, accounts []AccountInfo) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"alias_name", "account_id"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, account := range accounts {
		if err := writer.Write([]string{account.AliasName, account.AccountID}); err != nil {
			return fmt.Errorf("failed to write CSV data: %w", err)
		}
	}

	return nil
}