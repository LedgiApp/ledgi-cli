package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/LedgiApp/ledgi-cli/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with an API key",
	Long: `Save your Ledgi API key to ~/.ledgi/config.json.

The key must start with "ledgi_sk_". Get one from your Ledgi
dashboard under Settings > API Keys.

Examples:
  ledgi login --api-key ledgi_sk_abc123...
  ledgi login --api-key ledgi_sk_abc123... --api-url http://localhost:8000`,
	Run: runLogin,
}

var loginAPIKey string
var loginAPIURL string

func init() {
	loginCmd.Flags().StringVar(&loginAPIKey, "api-key", "", "API key (required, must start with ledgi_sk_)")
	loginCmd.Flags().StringVar(&loginAPIURL, "api-url", "", "API URL (optional, defaults to https://api.ledgi.uk)")
	loginCmd.MarkFlagRequired("api-key")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) {
	key := strings.TrimSpace(loginAPIKey)
	if !strings.HasPrefix(key, "ledgi_sk_") {
		fmt.Fprintln(os.Stderr, "Error: API key must start with 'ledgi_sk_'")
		os.Exit(1)
	}

	cfg := &config.Config{APIKey: key}
	if loginAPIURL != "" {
		cfg.APIURL = strings.TrimSpace(loginAPIURL)
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("API key saved to ~/.ledgi/config.json")
	fmt.Printf("Key prefix: %s...\n", key[:16])
}
