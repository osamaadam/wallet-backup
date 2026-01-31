# Wallet Backup - SMS Transaction Parser

A Go CLI tool that parses SMS backup files and automatically extracts bank transaction records into CSV files for expense tracking.

## What Does This Tool Do?

This tool converts SMS banking notifications from Egyptian banks (CIB and Banque Misr) into organized CSV expense records. It:

- **Parses SMS backups** in XML format (exported from Android SMS backup apps)
- **Extracts transaction details** including date, amount, payee, and transaction type
- **Automatically categorizes expenses** into predefined categories (Food, Shopping, Transportation, etc.)
- **Generates separate CSV files** for each account/card (current accounts, credit cards)
- **Supports multiple currencies** (EGP, USD, EUR, GBP, TRY, JPY)
- **Deduplicates transactions** to avoid double-counting
- **Cleans payee names** by removing payment processor prefixes

### Supported Banks

- **CIB (Commercial International Bank)**
  - Current/Debit accounts
  - Credit cards (automatically detects different cards by last 4 digits)
- **Banque Misr**
  - Current/Debit accounts

### Expense Categories

Transactions are automatically categorized into:

- Food & Drink
- Shopping
- Housing
- Transportation
- Vehicle
- Life & Entertainment
- Communication, PC
- Financial expenses
- Income

## Installation

### Prerequisites

- Go 1.16 or higher

### Install from Source

```bash
# Clone the repository
git clone https://github.com/osamaadam/wallet-backup.git
cd wallet-backup

# Install dependencies
go mod download

# Build the binary
go build -o sms-parser

# Optional: Install to your PATH
go install
```

### Quick Install (for Go users)

```bash
go install github.com/osamaadam/wallet-backup@latest
```

## Usage

### Basic Usage

```bash
# Parse SMS backup file (outputs CSV files to current directory)
./sms-parser sms-backup.xml
```

### Specify Output Directory

```bash
# Create CSV files in a specific directory
./sms-parser --output ./transactions sms-backup.xml

# Short form
./sms-parser -o ./output sms-backup.xml
```

The output directory will be automatically created if it doesn't exist.

### Getting Help

```bash
./sms-parser --help
```

## Output

The tool generates separate CSV files for each account/card:

- `CIB_Current_Debit.csv` - CIB debit card and current account transactions
- `CIB_Credit_Card_XXXX.csv` - CIB credit card transactions (one file per card)
- `Banque_Misr.csv` - Banque Misr account transactions

### CSV Format

Each CSV file contains the following columns (semicolon-delimited):

| Column   | Description                                    |
|----------|------------------------------------------------|
| date     | Transaction date and time (YYYY-MM-DD HH:MM:SS)|
| payee    | Merchant or transaction source                 |
| amount   | Transaction amount (negative for expenses)     |
| currency | Currency code (EGP, USD, EUR, etc.)           |
| type     | Transaction type (Expense or Income)           |
| category | Auto-assigned expense category                 |
| note     | Original SMS message with category prefix      |

The CSV files are UTF-8 encoded with BOM for proper display in Excel and other spreadsheet applications.

## How to Get SMS Backup

1. Use an Android SMS backup app (e.g., "SMS Backup & Restore")
2. Export your messages as XML
3. Transfer the XML file to your computer
4. Run this tool on the XML file

## Example

```bash
# Parse your SMS backup
./sms-parser --output ./my-expenses sms-20260131.xml

# Output:
# Created my-expenses/CIB_Current_Debit.csv with 195 transactions.
# Created my-expenses/Banque_Misr.csv with 394 transactions.
# Created my-expenses/CIB_Credit_Card_9018.csv with 1637 transactions.
# Created my-expenses/CIB_Credit_Card_5980.csv with 7 transactions.
```

## Development

### Run Without Building

```bash
go run main.go sms-backup.xml
```

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

Osama Adam - [GitHub](https://github.com/osamaadam)
