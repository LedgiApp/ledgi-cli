package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LedgiApp/ledgi-cli/internal/output"
	"github.com/spf13/cobra"
)

var isaCmd = &cobra.Command{
	Use:   "isa",
	Short: "View ISA summary and log deposits",
	Long: `View ISA allowance usage and log new deposits.
Tracks the UK £20,000 annual ISA allowance across all ISA accounts.`,
}

// --- summary ---

var isaSummaryTaxYear string

var isaSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show ISA allowance summary",
	Long: `Show ISA allowance usage for a given UK tax year (April 6 to April 5).

Scope required: isa:read

Options:
  --tax-year    Tax year start year, e.g. 2025 for 2025/26 (defaults to current)

Examples:
  ledgi isa summary
  ledgi isa summary --tax-year 2025
  ledgi isa summary --output table`,
	Run: runISASummary,
}

// --- deposit ---

var (
	isaDepositAccountID string
	isaDepositAmount    string
	isaDepositDate      string
)

var isaDepositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Log an ISA deposit",
	Long: `Log a deposit to an ISA account. The account must be an ISA type.
The deposit amount is validated against the remaining annual allowance.

Scope required: isa:write

Options:
  --account-id    Account ID (required)
  --amount        Deposit amount in GBP (required, must be > 0)
  --date          Deposit date YYYY-MM-DD (required)

Examples:
  ledgi isa deposit --account-id acc_123 --amount 500 --date 2026-02-24`,
	Run: runISADeposit,
}

func init() {
	isaSummaryCmd.Flags().StringVar(&isaSummaryTaxYear, "tax-year", "", "Tax year start (e.g. 2025)")

	isaDepositCmd.Flags().StringVar(&isaDepositAccountID, "account-id", "", "ISA account ID (required)")
	isaDepositCmd.Flags().StringVar(&isaDepositAmount, "amount", "", "Deposit amount (required)")
	isaDepositCmd.Flags().StringVar(&isaDepositDate, "date", "", "Deposit date YYYY-MM-DD (required)")
	isaDepositCmd.MarkFlagRequired("account-id")
	isaDepositCmd.MarkFlagRequired("amount")
	isaDepositCmd.MarkFlagRequired("date")

	isaCmd.AddCommand(isaSummaryCmd)
	isaCmd.AddCommand(isaDepositCmd)
	rootCmd.AddCommand(isaCmd)
}

func runISASummary(cmd *cobra.Command, args []string) {
	client := newClient()
	params := map[string]string{}
	if isaSummaryTaxYear != "" {
		params["tax_year"] = isaSummaryTaxYear
	}

	data, err := client.Get("/agent/isa/summary", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !isTable() {
		output.PrintJSON(data)
		return
	}

	var resp struct {
		TaxYear   string  `json:"tax_year"`
		Allowance float64 `json:"allowance"`
		Used      float64 `json:"used"`
		Remaining float64 `json:"remaining"`
		Currency  string  `json:"currency"`
		Deposits  []struct {
			AccountID string  `json:"account_id"`
			Amount    float64 `json:"amount"`
			Date      string  `json:"date"`
		} `json:"deposits"`
	}
	json.Unmarshal(data, &resp)

	fmt.Printf("Tax Year:     %s\n", resp.TaxYear)
	fmt.Printf("Allowance:    %s %.2f\n", resp.Currency, resp.Allowance)
	fmt.Printf("Used:         %s %.2f\n", resp.Currency, resp.Used)
	fmt.Printf("Remaining:    %s %.2f\n", resp.Currency, resp.Remaining)

	if len(resp.Deposits) > 0 {
		fmt.Println()
		headers := []string{"Date", "Account ID", "Amount"}
		var rows [][]string
		for _, d := range resp.Deposits {
			rows = append(rows, []string{
				d.Date, d.AccountID, fmt.Sprintf("%.2f", d.Amount),
			})
		}
		output.Table(headers, rows)
	}
}

func runISADeposit(cmd *cobra.Command, args []string) {
	client := newClient()

	body := map[string]any{
		"account_id": isaDepositAccountID,
		"amount":     json.Number(isaDepositAmount),
		"date":       isaDepositDate,
	}

	data, err := client.Post("/agent/isa/deposit", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output.PrintJSON(data)
}
