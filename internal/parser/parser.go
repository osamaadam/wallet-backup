package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"time"

	"sms-parser/internal/categorizer"
	"sms-parser/internal/models"
)

// Parser handles SMS backup parsing
type Parser struct {
	categorizer *categorizer.Categorizer
}

// New creates a new Parser instance
func New() *Parser {
	return &Parser{
		categorizer: categorizer.New(),
	}
}

// ParseFile reads and parses an SMS backup XML file with optional filters
func (p *Parser) ParseFile(filePath, senderFilter, startDateFilter string) (map[string][]models.Transaction, error) {
	// Read XML file
	xmlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Parse XML
	var backup models.SMSBackup
	if err := xml.Unmarshal(xmlFile, &backup); err != nil {
		return nil, fmt.Errorf("error parsing XML: %w", err)
	}

	// Parse start date filter if provided
	var startDate time.Time
	if startDateFilter != "" {
		startDate, err = time.Parse("2006-01-02", startDateFilter)
		if err != nil {
			return nil, fmt.Errorf("invalid date format (use YYYY-MM-DD): %w", err)
		}
	}

	// Initialize grouped data
	groupedData := map[string][]models.Transaction{
		"CIB_Current_Debit": {},
		"Banque_Misr":       {},
	}

	seenTransactions := make(map[string]bool)

	for _, sms := range backup.SMS {
		// Apply sender filter
		if senderFilter != "" && sms.Address != senderFilter {
			continue
		}

		// Create message signature for deduplication
		msgSignature := fmt.Sprintf("%s|%s|%s", sms.Date, sms.Address, sms.Body)
		if seenTransactions[msgSignature] {
			continue
		}
		seenTransactions[msgSignature] = true

		// Parse date
		dateMs, err := strconv.ParseInt(sms.Date, 10, 64)
		if err != nil {
			continue
		}
		dateObj := time.Unix(dateMs/1000, 0)

		// Apply date filter
		if !startDate.IsZero() && dateObj.Before(startDate) {
			continue
		}

		dateStr := dateObj.Format("2006-01-02 15:04:05")

		tx := models.Transaction{
			Date:     dateStr,
			Payee:    "",
			Amount:   0.0,
			Currency: "EGP",
			Type:     models.TypeExpense,
			Category: models.CatGeneral,
			Note:     sms.Body,
		}

		// Parse based on sender
		switch sms.Address {
		case "CIB":
			parseCIBMessage(&tx, sms.Body)
		case "Banque Misr":
			parseBanqueMisrMessage(&tx, sms.Body)
		}

		// Apply categorization
		if tx.TargetGroup != "" && tx.Amount != 0 && tx.Category == models.CatGeneral {
			tx.Category = p.categorizer.Categorize(tx.Payee, tx.Note, tx.Amount)
		}

		// Add category to note and append to group
		if tx.TargetGroup != "" && tx.Amount != 0 {
			if _, exists := groupedData[tx.TargetGroup]; !exists {
				groupedData[tx.TargetGroup] = []models.Transaction{}
			}

			if tx.Category != models.CatGeneral {
				tx.Note = fmt.Sprintf("[%s] %s", tx.Category, tx.Note)
			}

			groupedData[tx.TargetGroup] = append(groupedData[tx.TargetGroup], tx)
		}
	}

	return groupedData, nil
}
