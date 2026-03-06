package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LedgiApp/ledgi-cli/internal/output"
	"github.com/spf13/cobra"
)

var validAccountTypes = []string{
	"cash", "current", "savings", "credit_card", "investment",
	"pension", "property", "loan", "mortgage", "other_asset",
	"other_liability", "isa_cash", "isa_stocks", "isa_lifetime",
	"isa_innovative", "pension_workplace", "pension_sipp",
	"pension_state", "premium_bonds", "student_loan", "crypto_wallet",
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List and manage accounts",
	Long: `List and manage financial accounts (cash, ISAs, pensions, investments,
property, debt) via the Ledgi Agent API.`,
}

// --- list ---

var accountsListType string

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all accounts",
	Long: `List all active accounts for the authenticated user.

Scope required: accounts:read

Options:
  --type    Filter by account type (e.g. cash, isa_stocks, pension_sipp)

Examples:
  ledgi accounts list
  ledgi accounts list --type isa_stocks
  ledgi accounts list --type cash
  ledgi --output table accounts list`,
	Run: runAccountsList,
}

// --- upsert (single account) ---

var (
	upsertName        string
	upsertType        string
	upsertCurrency    string
	upsertBalance     float64
	upsertInstitution string
	upsertExternalID  string
)

var accountsUpsertSingleCmd = &cobra.Command{
	Use:   "upsert",
	Short: "Create or update a single account",
	Long: `Create or update a single account with inline flags.

Scope required: accounts:write

Required flags: --name, --type, --balance

Valid account types:
  Cash:       cash, current, savings, premium_bonds
  ISA:        isa_cash, isa_stocks, isa_lifetime, isa_innovative
  Pension:    pension, pension_workplace, pension_sipp, pension_state
  Investment: investment, crypto_wallet
  Property:   property
  Debt:       credit_card, loan, mortgage, student_loan
  Other:      other_asset, other_liability

Examples:
  ledgi accounts upsert --name "Main Current" --type current --currency GBP --balance 2500
  ledgi accounts upsert --name "Vanguard ISA" --type isa_stocks --balance 15000
  ledgi accounts upsert --name "Monzo" --type current --balance 1000 --institution Monzo --external-id monzo-1`,
	Run: runAccountsUpsertSingle,
}

// --- bulk-upsert ---

var accountsUpsertFile string

var accountsUpsertCmd = &cobra.Command{
	Use:   "bulk-upsert",
	Short: "Create or update accounts from a JSON file",
	Long: `Bulk create or update accounts by external_id (idempotent).

Scope required: accounts:write

The JSON file should contain:
{
  "accounts": [
    {
      "external_id": "my-current-1",
      "name": "Main Current Account",
      "type": "current",
      "balance": 2500.00,
      "currency": "GBP",
      "institution": "Monzo"
    }
  ]
}

Valid account types:
  cash, current, savings, credit_card, investment, pension, property,
  loan, mortgage, other_asset, other_liability, isa_cash, isa_stocks,
  isa_lifetime, isa_innovative, pension_workplace, pension_sipp,
  pension_state, premium_bonds, student_loan, crypto_wallet

Examples:
  ledgi accounts bulk-upsert --file accounts.json`,
	Run: runAccountsUpsert,
}

func init() {
	accountsListCmd.Flags().StringVar(&accountsListType, "type", "", "Filter by account type (e.g. cash, isa_stocks, current)")

	accountsUpsertSingleCmd.Flags().StringVar(&upsertName, "name", "", "Account name (required)")
	accountsUpsertSingleCmd.Flags().StringVar(&upsertType, "type", "", "Account type (required). See 'ledgi accounts upsert --help' for valid types")
	accountsUpsertSingleCmd.Flags().StringVar(&upsertCurrency, "currency", "GBP", "Currency code")
	accountsUpsertSingleCmd.Flags().Float64Var(&upsertBalance, "balance", 0, "Account balance (required)")
	accountsUpsertSingleCmd.Flags().StringVar(&upsertInstitution, "institution", "", "Institution name")
	accountsUpsertSingleCmd.Flags().StringVar(&upsertExternalID, "external-id", "", "External ID for idempotent upserts")
	accountsUpsertSingleCmd.MarkFlagRequired("name")
	accountsUpsertSingleCmd.MarkFlagRequired("type")

	accountsUpsertCmd.Flags().StringVar(&accountsUpsertFile, "file", "", "Path to JSON file (required)")
	accountsUpsertCmd.MarkFlagRequired("file")

	accountsCmd.AddCommand(accountsListCmd)
	accountsCmd.AddCommand(accountsUpsertSingleCmd)
	accountsCmd.AddCommand(accountsUpsertCmd)
	rootCmd.AddCommand(accountsCmd)
}

func runAccountsList(cmd *cobra.Command, args []string) {
	client := newClient()
	params := map[string]string{}
	if accountsListType != "" {
		params["type"] = accountsListType
	}

	data, err := client.Get("/agent/accounts", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !isTable() {
		output.PrintJSON(data)
		return
	}

	var resp struct {
		Accounts []struct {
			ID                string  `json:"id"`
			Name              string  `json:"name"`
			Type              string  `json:"type"`
			Balance           float64 `json:"balance"`
			Currency          string  `json:"currency"`
			Institution       *string `json:"institution"`
			IsLiability       bool    `json:"is_liability"`
			IncludeInNetWorth bool    `json:"include_in_net_worth"`
		} `json:"accounts"`
	}
	json.Unmarshal(data, &resp)

	headers := []string{"ID", "Name", "Type", "Balance", "Currency", "Institution"}
	var rows [][]string
	for _, a := range resp.Accounts {
		inst := ""
		if a.Institution != nil {
			inst = *a.Institution
		}
		rows = append(rows, []string{
			a.ID, a.Name, a.Type,
			fmt.Sprintf("%.2f", a.Balance),
			a.Currency, inst,
		})
	}
	output.Table(headers, rows)
}

func runAccountsUpsertSingle(cmd *cobra.Command, args []string) {
	client := newClient()

	account := map[string]any{
		"name":     upsertName,
		"type":     upsertType,
		"currency": upsertCurrency,
		"balance":  upsertBalance,
	}
	if upsertInstitution != "" {
		account["institution"] = upsertInstitution
	}
	if upsertExternalID != "" {
		account["external_id"] = upsertExternalID
	} else {
		// Auto-generate an external_id so the upsert endpoint works
		account["external_id"] = fmt.Sprintf("cli-%s-%s", upsertType, upsertName)
	}

	body := map[string]any{
		"accounts": []any{account},
	}

	resp, err := client.Post("/agent/accounts/bulk-upsert", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output.PrintJSON(resp)
}

func runAccountsUpsert(cmd *cobra.Command, args []string) {
	client := newClient()

	fileData, err := os.ReadFile(accountsUpsertFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var body json.RawMessage = fileData
	resp, err := client.Post("/agent/accounts/bulk-upsert", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output.PrintJSON(resp)
}
