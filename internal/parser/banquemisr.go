package parser

import (
	"regexp"
	"strconv"
	"strings"

	"sms-parser/internal/models"
	"sms-parser/internal/utils"
)

// parseBanqueMisrMessage parses Banque Misr bank SMS messages
func parseBanqueMisrMessage(tx *models.Transaction, body string) {
	tx.TargetGroup = "Banque_Misr"

	// Skip OTP and login messages
	skipWords := []string{"OTP", "password", "تسجيل الدخول", "code"}
	for _, word := range skipWords {
		if strings.Contains(body, word) {
			tx.TargetGroup = ""
			return
		}
	}

	if strings.Contains(body, "تم تحويل مبلغ") || strings.Contains(body, "تم اضافة مبلغ") {
		parseTransfer(tx, body)
	} else if strings.Contains(body, "تم الخصم") || strings.Contains(body, "transaction") {
		parsePurchase(tx, body)
	}
}

// parseTransfer handles Banque Misr transfer transactions
func parseTransfer(tx *models.Transaction, body string) {
	pattern := regexp.MustCompile(`مبلغ\s*(?:([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)\s*)?([\d,]+)(?:\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم))?`)
	match := pattern.FindStringSubmatch(body)

	if len(match) > 2 {
		val, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
		detectedCurr := match[1]
		if detectedCurr == "" {
			detectedCurr = match[3]
		}
		tx.Currency = utils.NormalizeCurrency(detectedCurr)

		if strings.Contains(body, "من حساب") {
			tx.Amount = -val
			tx.Payee = "Transfer Out"
		} else if strings.Contains(body, "الى حساب") {
			tx.Type = models.TypeIncome
			tx.Amount = val
			tx.Payee = "Transfer In"
		}
	}
}

// parsePurchase handles Banque Misr purchase transactions
func parsePurchase(tx *models.Transaction, body string) {
	pattern := regexp.MustCompile(`(?:مبلغ|amount)\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
	match := pattern.FindStringSubmatch(body)

	if len(match) > 2 {
		tx.Currency = utils.NormalizeCurrency(match[1])
		amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
		tx.Amount = -amount
		tx.Payee = "Card Purchase"

		tailPattern := regexp.MustCompile(`BM (.*?) (?:يوم|on)`)
		tailMatch := tailPattern.FindStringSubmatch(body)
		if len(tailMatch) > 1 {
			tx.Payee = strings.TrimSpace(tailMatch[1])
		}
	}
}
