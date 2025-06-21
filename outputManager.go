package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

type OutputManager interface {
	Output(accounts []AccountInfo, format string) error
}

type DefaultOutputManager struct{}

func NewOutputManager() OutputManager {
	return &DefaultOutputManager{}
}

func (om *DefaultOutputManager) Output(accounts []AccountInfo, format string) error {
	switch format {
	case "json":
		return om.outputJSON(accounts)
	case "table":
		return om.outputTable(accounts)
	case "csv":
		return om.outputCSV(accounts)
	case "standard", "":
		return om.outputStandard(accounts)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (om *DefaultOutputManager) outputJSON(accounts []AccountInfo) error {
	output := AccountInfoList{
		Accounts: accounts,
	}

	jsonData, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		return fmt.Errorf("error creating JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func (om *DefaultOutputManager) outputTable(accounts []AccountInfo) error {
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
			return fmt.Errorf("error appending table row: %w", err)
		}
	}

	table.Render()
	return nil
}

func (om *DefaultOutputManager) outputCSV(accounts []AccountInfo) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if err := writer.Write([]string{"id", "arn", "email", "name", "status", "joined_method", "joined_timestamp"}); err != nil {
		return fmt.Errorf("error writing CSV header: %w", err)
	}

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
			return fmt.Errorf("error writing CSV row: %w", err)
		}
	}

	return nil
}

func (om *DefaultOutputManager) outputStandard(accounts []AccountInfo) error {
	for _, account := range accounts {
		fmt.Printf("ID: %s | ARN: %s | Email: %s | Name: %s | Status: %s | Method: %s | Joined: %s\n",
			account.ID, account.Arn, account.Email, account.Name, account.Status, account.JoinedMethod, account.JoinedTimestamp)
	}
	return nil
}

func OutputAccounts(accounts []AccountInfo, format string) error {
	manager := NewOutputManager()
	return manager.Output(accounts, format)
}
