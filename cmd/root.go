package cmd

import (
	"fmt"
	"os"

	"github.com/LedgiApp/ledgi-cli/internal/api"
	"github.com/LedgiApp/ledgi-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	flagAPIKey string
	flagAPIURL string
	flagOutput string
)

var rootCmd = &cobra.Command{
	Use:   "ledgi",
	Short: "Ledgi CLI — Programmatic access to your Ledgi finance tracker",
	Long: `Ledgi CLI — Programmatic access to your Ledgi finance tracker.

Manage accounts, holdings, net worth snapshots, and ISA allowances
from the command line or any automation tool.

Setup:
  1. Get an API key from your Ledgi dashboard (Settings > API Keys)
  2. Run: ledgi login --api-key ledgi_sk_...
  3. Try: ledgi accounts list

Environment Variables:
  LEDGI_API_KEY   API key (overrides config file)
  LEDGI_API_URL   API URL (overrides config file)`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "Override API key for this request")
	rootCmd.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "Override API URL")
	rootCmd.PersistentFlags().StringVar(&flagOutput, "output", "json", "Output format: json (default), table")
}

// Execute is the main entry point for the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// newClient creates an API client using resolved config.
// It exits with an error message if no API key is configured.
func newClient() *api.Client {
	key, url := config.Resolve(flagAPIKey, flagAPIURL)
	if key == "" {
		fmt.Fprintln(os.Stderr, "Error: No API key configured.")
		fmt.Fprintln(os.Stderr, "Run: ledgi login --api-key ledgi_sk_...")
		fmt.Fprintln(os.Stderr, "Or set LEDGI_API_KEY environment variable.")
		os.Exit(1)
	}
	return api.New(key, url)
}

func isTable() bool {
	return flagOutput == "table"
}
