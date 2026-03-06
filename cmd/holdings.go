package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LedgiApp/ledgi-cli/internal/output"
	"github.com/spf13/cobra"
)

var holdingsCmd = &cobra.Command{
	Use:   "holdings",
	Short: "List and manage investment holdings",
	Long: `List and manage investment holdings (stocks, ETFs, crypto, bonds)
within your investment accounts via the Ledgi Agent API.`,
}

// --- list ---

var holdingsListAccountID string

var holdingsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all holdings",
	Long: `List all investment holdings for the authenticated user.

Scope required: holdings:read

Options:
  --account-id    Filter by parent account ID

Examples:
  ledgi holdings list
  ledgi holdings list --account-id acc_123
  ledgi holdings list --output table`,
	Run: runHoldingsList,
}

// --- bulk-upsert ---

var holdingsUpsertFile string

var holdingsUpsertCmd = &cobra.Command{
	Use:   "bulk-upsert",
	Short: "Create or update holdings from a JSON file",
	Long: `Bulk create or update holdings by external_id (idempotent).
Parent accounts are resolved by account_external_id.

Scope required: holdings:write

The JSON file should contain:
{
  "holdings": [
    {
      "external_id": "my-vwrl-1",
      "account_external_id": "my-isa-1",
      "symbol": "VWRL",
      "name": "Vanguard FTSE All-World",
      "type": "etf",
      "quantity": 50,
      "average_cost": 85.00,
      "currency": "GBP"
    }
  ]
}

Valid holding types: stock, etf, mutual_fund, bond, crypto, other

Examples:
  ledgi holdings bulk-upsert --file holdings.json`,
	Run: runHoldingsUpsert,
}

func init() {
	holdingsListCmd.Flags().StringVar(&holdingsListAccountID, "account-id", "", "Filter by parent account ID")
	holdingsUpsertCmd.Flags().StringVar(&holdingsUpsertFile, "file", "", "Path to JSON file (required)")
	holdingsUpsertCmd.MarkFlagRequired("file")

	holdingsCmd.AddCommand(holdingsListCmd)
	holdingsCmd.AddCommand(holdingsUpsertCmd)
	rootCmd.AddCommand(holdingsCmd)
}

func runHoldingsList(cmd *cobra.Command, args []string) {
	client := newClient()
	params := map[string]string{}
	if holdingsListAccountID != "" {
		params["account_id"] = holdingsListAccountID
	}

	data, err := client.Get("/agent/holdings", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !isTable() {
		output.PrintJSON(data)
		return
	}

	var resp struct {
		Holdings []struct {
			ID           string   `json:"id"`
			Symbol       string   `json:"symbol"`
			Name         string   `json:"name"`
			Type         string   `json:"type"`
			Quantity     float64  `json:"quantity"`
			AverageCost  float64  `json:"average_cost"`
			CurrentPrice *float64 `json:"current_price"`
			CurrentValue float64  `json:"current_value"`
			Currency     string   `json:"currency"`
		} `json:"holdings"`
	}
	json.Unmarshal(data, &resp)

	headers := []string{"Symbol", "Name", "Type", "Qty", "Avg Cost", "Price", "Value", "Currency"}
	var rows [][]string
	for _, h := range resp.Holdings {
		price := "—"
		if h.CurrentPrice != nil {
			price = fmt.Sprintf("%.2f", *h.CurrentPrice)
		}
		rows = append(rows, []string{
			h.Symbol, h.Name, h.Type,
			fmt.Sprintf("%.4f", h.Quantity),
			fmt.Sprintf("%.2f", h.AverageCost),
			price,
			fmt.Sprintf("%.2f", h.CurrentValue),
			h.Currency,
		})
	}
	output.Table(headers, rows)
}

func runHoldingsUpsert(cmd *cobra.Command, args []string) {
	client := newClient()

	fileData, err := os.ReadFile(holdingsUpsertFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var body json.RawMessage = fileData
	resp, err := client.Post("/agent/holdings/bulk-upsert", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output.PrintJSON(resp)
}
