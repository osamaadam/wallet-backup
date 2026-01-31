package cmd

import (
	"fmt"
	"os"

	"sms-parser/internal/parser"
	"sms-parser/internal/writer"

	"github.com/spf13/cobra"
)

var (
	outputDir  string
	senderName string
	startDate  string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sms-parser [xml-file]",
	Short: "Parse SMS backup and extract bank transactions",
	Long:  `A CLI tool to parse SMS backup XML files and extract bank transactions into CSV files.`,
	Args:  cobra.ExactArgs(1),
	RunE:  run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for CSV files (created if not exists)")
	RootCmd.Flags().StringVarP(&senderName, "sender", "s", "", "Filter by sender name (e.g., 'CIB', 'Banque Misr')")
	RootCmd.Flags().StringVarP(&startDate, "from", "f", "", "Filter messages from this date onwards (format: YYYY-MM-DD)")
}

func run(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parse the SMS backup file
	p := parser.New()
	transactions, err := p.ParseFile(filePath, senderName, startDate)
	if err != nil {
		return fmt.Errorf("failed to parse SMS backup: %w", err)
	}

	// Write transactions to CSV files
	w := writer.New(outputDir)
	if err := w.Write(transactions); err != nil {
		return fmt.Errorf("failed to write transactions: %w", err)
	}

	return nil
}
