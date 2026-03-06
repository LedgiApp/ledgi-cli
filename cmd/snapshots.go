package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/LedgiApp/ledgi-cli/internal/output"
	"github.com/spf13/cobra"
)

var snapshotsCmd = &cobra.Command{
	Use:   "snapshots",
	Short: "List and create net worth snapshots",
	Long: `View historical net worth snapshots and create new ones.
Snapshots capture a point-in-time record of all accounts and holdings.`,
}

// --- list ---

var snapshotsListLimit string

var snapshotsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List historical snapshots",
	Long: `List net worth snapshots in chronological order.

Scope required: snapshots:read

Options:
  --limit    Maximum number of snapshots to return (default: 12, max: 100)

Examples:
  ledgi snapshots list
  ledgi snapshots list --limit 6
  ledgi snapshots list --output table`,
	Run: runSnapshotsList,
}

// --- create ---

var snapshotsCreateDate string

var snapshotsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new net worth snapshot",
	Long: `Create a point-in-time net worth snapshot from current account and holdings data.

Scope required: snapshots:write

Options:
  --date    Snapshot date in YYYY-MM-DD format (defaults to today)

Examples:
  ledgi snapshots create
  ledgi snapshots create --date 2026-02-01`,
	Run: runSnapshotsCreate,
}

func init() {
	snapshotsListCmd.Flags().StringVar(&snapshotsListLimit, "limit", "12", "Max snapshots to return (1-100)")
	snapshotsCreateCmd.Flags().StringVar(&snapshotsCreateDate, "date", "", "Snapshot date (YYYY-MM-DD, defaults to today)")

	snapshotsCmd.AddCommand(snapshotsListCmd)
	snapshotsCmd.AddCommand(snapshotsCreateCmd)
	rootCmd.AddCommand(snapshotsCmd)
}

func runSnapshotsList(cmd *cobra.Command, args []string) {
	client := newClient()
	params := map[string]string{
		"limit": snapshotsListLimit,
	}

	data, err := client.Get("/agent/snapshots", params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !isTable() {
		output.PrintJSON(data)
		return
	}

	var resp struct {
		Snapshots []struct {
			ID               string  `json:"id"`
			Date             string  `json:"date"`
			NetWorth         float64 `json:"net_worth"`
			TotalAssets      float64 `json:"total_assets"`
			TotalLiabilities float64 `json:"total_liabilities"`
		} `json:"snapshots"`
	}
	json.Unmarshal(data, &resp)

	headers := []string{"Date", "Net Worth", "Assets", "Liabilities"}
	var rows [][]string
	for _, s := range resp.Snapshots {
		rows = append(rows, []string{
			s.Date,
			fmt.Sprintf("%.2f", s.NetWorth),
			fmt.Sprintf("%.2f", s.TotalAssets),
			fmt.Sprintf("%.2f", s.TotalLiabilities),
		})
	}
	output.Table(headers, rows)
}

func runSnapshotsCreate(cmd *cobra.Command, args []string) {
	client := newClient()

	body := map[string]any{}
	if snapshotsCreateDate != "" {
		body["snapshot_date"] = snapshotsCreateDate
	}

	data, err := client.Post("/agent/snapshots", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output.PrintJSON(data)
}
