package writer

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"sms-parser/internal/models"
)

// Writer handles CSV file writing
type Writer struct {
	outputDir string
}

// New creates a new Writer instance
func New(outputDir string) *Writer {
	return &Writer{
		outputDir: outputDir,
	}
}

// Write writes transactions to CSV files grouped by account
func (w *Writer) Write(groupedData map[string][]models.Transaction) error {
	fieldnames := []string{"date", "payee", "amount", "currency", "type", "category", "note"}

	for groupName, transactions := range groupedData {
		if len(transactions) == 0 {
			continue
		}

		// Sort by date
		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].Date < transactions[j].Date
		})

		// Create CSV file
		filename := filepath.Join(w.outputDir, groupName+".csv")
		if err := w.writeCSVFile(filename, fieldnames, transactions); err != nil {
			return err
		}

		fmt.Printf("Created %s with %d transactions.\n", filename, len(transactions))
	}

	return nil
}

// writeCSVFile writes a single CSV file
func (w *Writer) writeCSVFile(filename string, headers []string, transactions []models.Transaction) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", filename, err)
	}
	defer file.Close()

	// Write BOM for UTF-8
	if _, err := file.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return fmt.Errorf("error writing BOM to %s: %w", filename, err)
	}

	writer := csv.NewWriter(file)
	writer.Comma = ';'

	// Write header
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error writing header to %s: %w", filename, err)
	}

	// Write transactions
	for _, tx := range transactions {
		record := []string{
			tx.Date,
			tx.Payee,
			fmt.Sprintf("%.2f", tx.Amount),
			tx.Currency,
			tx.Type,
			tx.Category,
			tx.Note,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writing transaction to %s: %w", filename, err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing writer for %s: %w", filename, err)
	}

	return nil
}
