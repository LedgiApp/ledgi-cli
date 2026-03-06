# Ledgi CLI

Command-line interface for the [Ledgi](https://ledgi.app) Agent API. Manage your accounts, holdings, net worth, snapshots, and ISA deposits from the terminal.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/LedgiApp/ledgi-cli/main/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/LedgiApp/ledgi-cli.git
cd ledgi-cli
go build -o ledgi .
```

## Quick start

```bash
# Authenticate with your API key (Settings > API Keys in the Ledgi dashboard)
ledgi login --api-key ledgi_sk_...

# List your accounts
ledgi accounts list

# Check your net worth
ledgi networth summary
```

## Commands

### Authentication

```bash
ledgi login --api-key ledgi_sk_...
ledgi login --api-key ledgi_sk_... --api-url https://custom-url.com
```

Credentials are saved to `~/.ledgi/config.json` (file permissions `0600`).

### Accounts

```bash
# List all accounts
ledgi accounts list

# Filter by type
ledgi accounts list --type cash
ledgi accounts list --type isa_stocks

# Create or update a single account
ledgi accounts upsert \
  --name "Main Current" \
  --type current \
  --currency GBP \
  --balance 2500 \
  --institution Monzo \
  --external-id monzo-1

# Bulk upsert from a JSON file
ledgi accounts bulk-upsert --file accounts.json
```

**Valid account types:**

| Category | Types |
|----------|-------|
| Cash | `cash`, `current`, `savings`, `premium_bonds` |
| ISA | `isa_cash`, `isa_stocks`, `isa_lifetime`, `isa_innovative` |
| Pension | `pension`, `pension_workplace`, `pension_sipp`, `pension_state` |
| Investment | `investment`, `crypto_wallet` |
| Property | `property` |
| Debt | `credit_card`, `loan`, `mortgage`, `student_loan` |
| Other | `other_asset`, `other_liability` |

### Holdings

```bash
# List all holdings
ledgi holdings list

# Filter by account
ledgi holdings list --account-id abc123

# Bulk upsert from a JSON file
ledgi holdings bulk-upsert --file holdings.json
```

### Net worth

```bash
ledgi networth summary
```

### Snapshots

```bash
# List recent snapshots
ledgi snapshots list
ledgi snapshots list --limit 5

# Create a snapshot
ledgi snapshots create
ledgi snapshots create --date 2026-01-31
```

### ISA

```bash
# View ISA summary
ledgi isa summary
ledgi isa summary --tax-year 2025

# Log a deposit
ledgi isa deposit --account-id abc123 --amount 1000 --date 2026-02-24
```

## Output formats

All commands default to JSON output. Use `--output table` for a human-readable table:

```bash
ledgi --output table accounts list
```

## Configuration

The CLI resolves configuration in this order (highest priority first):

1. **Flags** — `--api-key`, `--api-url`
2. **Environment variables** — `LEDGI_API_KEY`, `LEDGI_API_URL`
3. **Config file** — `~/.ledgi/config.json`
4. **Default** — API URL defaults to `https://api.ledgi.app`

## API key scopes

The CLI requires an API key with the appropriate scopes. Create one from **Settings > API Keys** in the Ledgi dashboard.

| Scope | Commands |
|-------|----------|
| `accounts:read` | `accounts list` |
| `accounts:write` | `accounts upsert`, `accounts bulk-upsert` |
| `holdings:read` | `holdings list` |
| `holdings:write` | `holdings bulk-upsert` |
| `networth:read` | `networth summary` |
| `snapshots:read` | `snapshots list` |
| `snapshots:write` | `snapshots create` |
| `isa:read` | `isa summary` |
| `isa:write` | `isa deposit` |

## License

Copyright Ledgi. All rights reserved.
