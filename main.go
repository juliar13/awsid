package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type AccountInfo struct {
	ID            string `json:"id"`
	Arn           string `json:"arn"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	JoinedMethod  string `json:"joined_method"`
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
	var formatOption string
	var sortField string
	var sortDesc string
	var rootCmd = &cobra.Command{
		Use:     "awsid [alias_name]",
		Short:   "Get AWS account ID from alias name",
		Long:    "A CLI tool to get AWS account ID from alias name. Supports both positional arguments and --name option.",
		Version: Version,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// Validate and resolve format flags
			resolvedFormat, err := resolveFormatFlags(formatOption, jsonOutput, tableOutput, csvOutput)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			
			// Validate and resolve sort flags
			resolvedSort, err := resolveSortFlags(sortField, sortDesc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
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

				// If exact match found
				if len(exactMatch) > 0 {
					sortAccounts(exactMatch, resolvedSort)
					outputByFormat(exactMatch, resolvedFormat, true)
					return
				}

				// If partial matches found
				if len(matchingAccounts) > 0 {
					sortAccounts(matchingAccounts, resolvedSort)
					outputByFormat(matchingAccounts, resolvedFormat, false)
					return
				}

				// No matches found
				fmt.Fprintf(os.Stderr, "No account found with alias name: %s\n", searchTerm)
				os.Exit(1)
			} else {
				// No search term provided, list all accounts
				sortAccounts(accounts, resolvedSort)
				outputByFormat(accounts, resolvedFormat, false)
			}
		},
	}

	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.Flags().BoolVar(&tableOutput, "table", false, "Output in table format")
	rootCmd.Flags().BoolVar(&csvOutput, "csv", false, "Output in CSV format")
	rootCmd.Flags().StringVar(&formatOption, "format", "", "Output format (json, table, csv)")
	rootCmd.Flags().StringVar(&nameSearch, "name", "", "Search by account name (takes priority over positional argument)")
	rootCmd.Flags().StringVar(&sortField, "sort", "", "Sort by field (id, name, email, status, joined_timestamp, joined_method)")
	rootCmd.Flags().StringVar(&sortDesc, "sort-desc", "", "Sort by field in descending order (id, name, email, status, joined_timestamp, joined_method)")


	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// resolveFormatFlags resolves format conflicts and determines final format
func resolveFormatFlags(formatOption string, jsonOutput, tableOutput, csvOutput bool) (string, error) {
	// Count active format flags
	activeFlags := 0
	if jsonOutput {
		activeFlags++
	}
	if tableOutput {
		activeFlags++
	}
	if csvOutput {
		activeFlags++
	}
	
	// Check for multiple individual format flags
	if activeFlags > 1 {
		return "", fmt.Errorf("multiple output format flags specified. Use only one format option")
	}
	
	// If --format is specified, validate and use it (takes priority)
	if formatOption != "" {
		if err := validateFormat(formatOption); err != nil {
			return "", err
		}
		return formatOption, nil
	}
	
	// If individual format flag is specified, use it
	if jsonOutput {
		return "json", nil
	}
	if tableOutput {
		return "table", nil
	}
	if csvOutput {
		return "csv", nil
	}
	
	// Default format (no flags specified - backward compatible behavior)
	return "default", nil
}

// validateFormat validates the format string
func validateFormat(format string) error {
	if format == "" {
		return fmt.Errorf("output format cannot be empty. Supported formats: json, table, csv")
	}
	
	validFormats := []string{"json", "table", "csv"}
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid output format \"%s\". Supported formats: json, table, csv", format)
}

// SortInfo holds sort configuration
type SortInfo struct {
	Field      string
	Descending bool
}

// resolveSortFlags validates and resolves sort configuration
func resolveSortFlags(sortField, sortDesc string) (*SortInfo, error) {
	// Check for conflicting sort flags
	if sortField != "" && sortDesc != "" {
		return nil, fmt.Errorf("cannot specify both --sort and --sort-desc. Use only one sort option")
	}
	
	// No sort specified
	if sortField == "" && sortDesc == "" {
		return &SortInfo{}, nil
	}
	
	// Determine field and direction
	var field string
	var desc bool
	
	if sortField != "" {
		field = sortField
		desc = false
	} else {
		field = sortDesc
		desc = true
	}
	
	// Validate sort field
	if err := validateSortField(field); err != nil {
		return nil, err
	}
	
	return &SortInfo{Field: field, Descending: desc}, nil
}

// validateSortField validates the sort field name
func validateSortField(field string) error {
	validFields := []string{"id", "name", "email", "status", "joined_timestamp", "joined_method"}
	for _, valid := range validFields {
		if field == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid sort field \"%s\". Supported fields: id, name, email, status, joined_timestamp, joined_method", field)
}

// sortAccounts sorts accounts based on the provided sort configuration
func sortAccounts(accounts []AccountInfo, sortInfo *SortInfo) {
	if sortInfo.Field == "" {
		return // No sorting required
	}
	
	sort.Slice(accounts, func(i, j int) bool {
		var result bool
		
		switch sortInfo.Field {
		case "id":
			result = accounts[i].ID < accounts[j].ID
		case "name":
			result = strings.ToLower(accounts[i].Name) < strings.ToLower(accounts[j].Name)
		case "email":
			result = strings.ToLower(accounts[i].Email) < strings.ToLower(accounts[j].Email)
		case "status":
			result = accounts[i].Status < accounts[j].Status
		case "joined_timestamp":
			result = accounts[i].JoinedTimestamp < accounts[j].JoinedTimestamp
		case "joined_method":
			result = accounts[i].JoinedMethod < accounts[j].JoinedMethod
		default:
			return false // Should not happen due to validation
		}
		
		// Reverse for descending order
		if sortInfo.Descending {
			result = !result
		}
		
		return result
	})
}

// outputByFormat outputs accounts using the specified format
func outputByFormat(accounts []AccountInfo, format string, isExactMatch bool) {
	switch format {
	case "json":
		outputJSON(accounts)
	case "table":
		outputTable(accounts)
	case "csv":
		outputCSV(accounts)
	case "default":
		// Default format: show account IDs for exact matches, detailed info for partial matches
		if isExactMatch && len(accounts) > 0 {
			fmt.Println(accounts[0].AccountID)
		} else {
			for _, account := range accounts {
				fmt.Printf("ID: %s | ARN: %s | Email: %s | Name: %s | Status: %s | Method: %s | Joined: %s\n", 
					account.ID, account.Arn, account.Email, account.Name, account.Status, account.JoinedMethod, account.JoinedTimestamp)
			}
		}
	default:
		// Fallback to table format
		outputTable(accounts)
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
					ID:            strings.TrimSpace(record[0]),
					Arn:           strings.TrimSpace(record[1]),
					Email:         strings.TrimSpace(record[2]),
					Name:          strings.TrimSpace(record[3]),
					Status:        strings.TrimSpace(record[4]),
					JoinedMethod:  strings.TrimSpace(record[5]),
					JoinedTimestamp: strings.TrimSpace(record[6]),
					// Backward compatibility
					AliasName:     strings.TrimSpace(record[3]), // Name -> AliasName
					AccountID:     strings.TrimSpace(record[0]), // ID -> AccountID
				}
			} else if len(record) >= 2 {
				// Old format: alias_name, account_id
				account = AccountInfo{
					ID:            strings.TrimSpace(record[1]), // account_id -> ID
					Name:          strings.TrimSpace(record[0]), // alias_name -> Name
					AliasName:     strings.TrimSpace(record[0]),
					AccountID:     strings.TrimSpace(record[1]),
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
				ID:     *account.Id,
				Name:   *account.Name,
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