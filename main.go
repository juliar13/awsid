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
	ID              string `json:"id"`
	Arn             string `json:"arn"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	JoinedMethod    string `json:"joined_method"`
	JoinedTimestamp string `json:"joined_timestamp"`
	// Backward compatibility fields
	AliasName string `json:"alias_name"`
	AccountID string `json:"account_id"`
}

type AccountInfoList struct {
	Accounts []AccountInfo `json:"account_info"`
}

const Version = "0.5.0"

func main() {
	var jsonOutput bool
	var tableOutput bool
	var csvOutput bool
	var nameSearch string
	var rootCmd = &cobra.Command{
		Use:     "awsid [alias_name]",
		Short:   "Get AWS account ID from alias name",
		Long:    "A CLI tool to get AWS account ID from alias name. Supports both positional arguments and --name option.",
		Version: Version,
		Args:    cobra.MinimumNArgs(0),
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

			// Determine search term: --name option takes priority over positional argument
			var searchTerm string
			if nameSearch != "" {
				searchTerm = nameSearch
			} else if len(args) > 0 {
				searchTerm = args[0]
			}

			// Determine output format
			format := getOutputFormat(jsonOutput, tableOutput, csvOutput)

			// If search term is provided, search for matching accounts
			if searchTerm != "" {
				matchingAccounts := []AccountInfo{}

				for _, account := range accounts {
					if strings.Contains(account.AliasName, searchTerm) {
						matchingAccounts = append(matchingAccounts, account)
					}
				}

				// Check for exact match first
				exactMatch := []AccountInfo{}
				for _, account := range matchingAccounts {
					if account.AliasName == searchTerm {
						exactMatch = append(exactMatch, account)
						break
					}
				}

				// If exact match found and not JSON output, show only account ID for standard format
				if len(exactMatch) > 0 {
					if format == "standard" {
						fmt.Println(exactMatch[0].AccountID)
						return
					} else {
						if err := OutputAccounts(exactMatch, format); err != nil {
							fmt.Fprintf(os.Stderr, "Error outputting accounts: %v\n", err)
							os.Exit(1)
						}
						return
					}
				}

				// If partial matches found, output them
				if len(matchingAccounts) > 0 {
					if err := OutputAccounts(matchingAccounts, format); err != nil {
						fmt.Fprintf(os.Stderr, "Error outputting accounts: %v\n", err)
						os.Exit(1)
					}
					return
				}

				// No matches found
				fmt.Fprintf(os.Stderr, "No account found with alias name: %s\n", searchTerm)
				os.Exit(1)
			} else {
				// No search term provided, list all accounts
				if err := OutputAccounts(accounts, format); err != nil {
					fmt.Fprintf(os.Stderr, "Error outputting accounts: %v\n", err)
					os.Exit(1)
				}
			}
		},
	}

	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&tableOutput, "table", false, "Output in table format")
	rootCmd.Flags().BoolVar(&csvOutput, "csv", false, "Output in CSV format")
	rootCmd.Flags().StringVar(&nameSearch, "name", "", "Search by account name (takes priority over positional argument)")

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
		if i == 0 && (len(record) > 0 && (record[0] == "alias_name" || record[0] == "AliasName" || record[0] == "id")) {
			continue
		}

		if len(record) >= 2 && record[0] != "" {
			var account AccountInfo

			// Check if this is the new format (7 columns) or old format (2 columns)
			if len(record) >= 7 {
				// New format: id, arn, email, name, status, joined_method, joined_timestamp
				account = AccountInfo{
					ID:              strings.TrimSpace(record[0]),
					Arn:             strings.TrimSpace(record[1]),
					Email:           strings.TrimSpace(record[2]),
					Name:            strings.TrimSpace(record[3]),
					Status:          strings.TrimSpace(record[4]),
					JoinedMethod:    strings.TrimSpace(record[5]),
					JoinedTimestamp: strings.TrimSpace(record[6]),
					// Backward compatibility
					AliasName: strings.TrimSpace(record[3]), // Name -> AliasName
					AccountID: strings.TrimSpace(record[0]), // ID -> AccountID
				}
			} else if len(record) >= 2 {
				// Old format: alias_name, account_id
				account = AccountInfo{
					ID:        strings.TrimSpace(record[1]), // account_id -> ID
					Name:      strings.TrimSpace(record[0]), // alias_name -> Name
					AliasName: strings.TrimSpace(record[0]),
					AccountID: strings.TrimSpace(record[1]),
				}
			}

			if account.ID != "" {
				accounts = append(accounts, account)
			}
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
	table.Header("ID", "ARN", "Email", "Name", "Status", "Joined Method", "Joined Timestamp")

	for _, account := range accounts {
		err := table.Append([]any{
			account.ID,
			account.Arn,
			account.Email,
			account.Name,
			account.Status,
			account.JoinedMethod,
			account.JoinedTimestamp,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error appending table row: %v\n", err)
			continue
		}
	}

	table.Render()
}
func outputCSV(accounts []AccountInfo) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "arn", "email", "name", "status", "joined_method", "joined_timestamp"}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV header: %v\n", err)
		return
	}

	// Write data
	for _, account := range accounts {
		if err := writer.Write([]string{
			account.ID,
			account.Arn,
			account.Email,
			account.Name,
			account.Status,
			account.JoinedMethod,
			account.JoinedTimestamp,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing CSV row: %v\n", err)
			continue
		}
	}
}

func updateAccountInfoFromAWS(filePath string) error {
	// Create .aws directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Load AWS configuration with us-east-1 region (Organizations is global but requires a region)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
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
			accountInfo := AccountInfo{
				ID:   *account.Id,
				Name: *account.Name,
				// Backward compatibility
				AliasName: *account.Name,
				AccountID: *account.Id,
			}

			if account.Arn != nil {
				accountInfo.Arn = *account.Arn
			}
			if account.Email != nil {
				accountInfo.Email = *account.Email
			}
			accountInfo.Status = string(account.Status)
			accountInfo.JoinedMethod = string(account.JoinedMethod)
			if account.JoinedTimestamp != nil {
				accountInfo.JoinedTimestamp = account.JoinedTimestamp.Format("2006-01-02T15:04:05.000000-07:00")
			}

			accounts = append(accounts, accountInfo)
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
	if err := writer.Write([]string{"id", "arn", "email", "name", "status", "joined_method", "joined_timestamp"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, account := range accounts {
		if err := writer.Write([]string{
			account.ID,
			account.Arn,
			account.Email,
			account.Name,
			account.Status,
			account.JoinedMethod,
			account.JoinedTimestamp,
		}); err != nil {
			return fmt.Errorf("failed to write CSV data: %w", err)
		}
	}

	return nil
}

func getOutputFormat(jsonOutput, tableOutput, csvOutput bool) string {
	if jsonOutput {
		return "json"
	}
	if tableOutput {
		return "table"
	}
	if csvOutput {
		return "csv"
	}
	return "standard"
}
