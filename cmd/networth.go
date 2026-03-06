package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LedgiApp/ledgi-cli/internal/output"
	"github.com/spf13/cobra"
)

var networthCmd = &cobra.Command{
	Use:   "networth",
	Short: "View net worth summary and breakdown",
	Long:  `View your current net worth calculated from all accounts and holdings.`,
}

var networthSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show current net worth",
	Long: `Calculate and display the current net worth with a breakdown by account type.

Scope required: networth:read

The calculation includes all accounts marked "include_in_net_worth" and
all holdings with non-zero values. Prices use the last cached values
(not live-fetched).

Examples:
  ledgi networth summary
  ledgi networth summary --output table`,
	Run: runNetworthSummary,
}

func init() {
	networthCmd.AddCommand(networthSummaryCmd)
	rootCmd.AddCommand(networthCmd)
}

func runNetworthSummary(cmd *cobra.Command, args []string) {
	client := newClient()

	data, err := client.Get("/agent/networth", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !isTable() {
		output.PrintJSON(data)
		return
	}

	var resp struct {
		TotalNetWorth    float64            `json:"total_net_worth"`
		TotalAssets      float64            `json:"total_assets"`
		TotalLiabilities float64            `json:"total_liabilities"`
		Currency         string             `json:"currency"`
		Breakdown        map[string]float64 `json:"breakdown"`
	}
	json.Unmarshal(data, &resp)

	fmt.Printf("Net Worth:    %s %.2f\n", resp.Currency, resp.TotalNetWorth)
	fmt.Printf("Assets:       %s %.2f\n", resp.Currency, resp.TotalAssets)
	fmt.Printf("Liabilities:  %s %.2f\n", resp.Currency, resp.TotalLiabilities)
	fmt.Println()

	if len(resp.Breakdown) > 0 {
		headers := []string{"Category", "Amount"}
		var rows [][]string
		for k, v := range resp.Breakdown {
			rows = append(rows, []string{k, fmt.Sprintf("%.2f", v)})
		}
		output.Table(headers, rows)
	}
}
