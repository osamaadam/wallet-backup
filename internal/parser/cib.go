package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"sms-parser/internal/models"
	"sms-parser/internal/utils"
)

// parseCIBMessage parses CIB bank SMS messages
func parseCIBMessage(tx *models.Transaction, body string) {
	// Detect credit card
	ccPattern := regexp.MustCompile(`(?i)(?:credit card|ending with|card|بـ)\s*[#*]*\s*(\d{4})`)
	ccMatch := ccPattern.FindStringSubmatch(body)

	isCreditCard := false
	cardDigits := "Unknown"

	if len(ccMatch) > 1 {
		cardDigits = ccMatch[1]
		// If it's not the Debit Card (7759) and not the Account (2373)
		if cardDigits != "7759" && cardDigits != "2373" {
			isCreditCard = true
			tx.TargetGroup = fmt.Sprintf("CIB_Credit_Card_%s", cardDigits)
		}
	}

	if isCreditCard {
		parseCIBCreditCard(tx, body)
	} else if strings.Contains(body, "7759") || strings.Contains(body, "2373") {
		parseCIBDebit(tx, body)
	}
}

// parseCIBCreditCard handles CIB credit card transactions
func parseCIBCreditCard(tx *models.Transaction, body string) {
	if strings.Contains(body, "charged for") || strings.Contains(body, "purchasing transaction") {
		pattern := regexp.MustCompile(`(?i)charged for\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})\s*at\s*(.*?)(?:\s+on|\s+at|\. Available)`)
		match := pattern.FindStringSubmatch(body)
		if len(match) > 3 {
			tx.Currency = utils.NormalizeCurrency(match[1])
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
			tx.Amount = -amount
			tx.Payee = utils.CleanPayeeName(strings.TrimSpace(match[3]))
		}
	} else if strings.Contains(body, "refunded") || strings.Contains(body, "rad") || strings.Contains(body, "رد") {
		if !strings.Contains(body, "تم سداد") {
			tx.Type = models.TypeIncome
			pattern := regexp.MustCompile(`(?i)(?:refunded|red|rd|رد)\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
			match := pattern.FindStringSubmatch(body)
			if len(match) > 2 {
				tx.Currency = utils.NormalizeCurrency(match[1])
				amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
				tx.Amount = amount
				tx.Payee = "Refund"
			}
		}
	}

	if strings.Contains(body, "تم سداد") || (strings.Contains(body, "payment") && strings.Contains(body, "received")) {
		tx.Type = models.TypeIncome
		tx.Payee = "CIB Repayment"
		pattern := regexp.MustCompile(`مبلغ\s*([\d,]+\.\d{2})`)
		match := pattern.FindStringSubmatch(body)
		if len(match) > 1 {
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
			tx.Amount = amount
		}
	}
}

// parseCIBDebit handles CIB debit card and current account transactions
func parseCIBDebit(tx *models.Transaction, body string) {
	tx.TargetGroup = "CIB_Current_Debit"

	if strings.Contains(body, "7759") &&
		(strings.Contains(body, "charged for") || strings.Contains(body, "خصم") ||
			strings.Contains(body, "withdrawal") || strings.Contains(body, "سحب")) {

		// Arabic pattern
		patternAr := regexp.MustCompile(`خصم\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})\s*من.*?عند\s*(.*?)(\s+في|$)`)
		matchAr := patternAr.FindStringSubmatch(body)

		// English pattern
		patternEn := regexp.MustCompile(`(?i)charged for\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})\s*at\s*(.*?)(?:\s+on|\s+at)`)
		matchEn := patternEn.FindStringSubmatch(body)

		// Withdrawal pattern
		patternWith := regexp.MustCompile(`سحب\s*(?:مبلغ)?\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
		matchWith := patternWith.FindStringSubmatch(body)

		if len(matchAr) > 3 {
			tx.Currency = utils.NormalizeCurrency(matchAr[1])
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchAr[2], ",", ""), 64)
			tx.Amount = -amount
			tx.Payee = utils.CleanPayeeName(strings.TrimSpace(matchAr[3]))
		} else if len(matchEn) > 3 {
			tx.Currency = utils.NormalizeCurrency(matchEn[1])
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchEn[2], ",", ""), 64)
			tx.Amount = -amount
			tx.Payee = utils.CleanPayeeName(strings.TrimSpace(matchEn[3]))
		} else if len(matchWith) > 2 {
			tx.Currency = utils.NormalizeCurrency(matchWith[1])
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchWith[2], ",", ""), 64)
			tx.Amount = -amount
			tx.Payee = "ATM Withdrawal"
		}
	} else if strings.Contains(body, "2373") {
		parseCIBCurrentAccount(tx, body)
	}
}

// parseCIBCurrentAccount handles CIB current account transactions
func parseCIBCurrentAccount(tx *models.Transaction, body string) {
	if strings.Contains(body, "debited") || strings.Contains(body, "charged with") || strings.Contains(body, "تم تحويل") {
		pattern := regexp.MustCompile(`(?i)(?:amount|for)\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
		match := pattern.FindStringSubmatch(body)
		if len(match) > 2 {
			tx.Currency = utils.NormalizeCurrency(match[1])
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
			tx.Amount = -amount

			if strings.Contains(body, "transfer to another account") {
				tx.Payee = "Transfer to Account / CC"
				tx.Category = models.CatFinancial
			} else {
				payeePattern := regexp.MustCompile(`to\s+(.*?)\s+with reference`)
				payeeMatch := payeePattern.FindStringSubmatch(body)
				if len(payeeMatch) > 1 {
					tx.Payee = strings.TrimSpace(payeeMatch[1])
				} else {
					tx.Payee = "Transfer Out"
				}
			}
		}
	} else if strings.Contains(body, "credited") || strings.Contains(body, "تحويل مبلغ") || strings.Contains(body, "add") {
		tx.Type = models.TypeIncome

		// IPN pattern
		patternIPN := regexp.MustCompile(`(?i)credited with IPN Inward for\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
		matchIPN := patternIPN.FindStringSubmatch(body)

		// Salary pattern
		patternSal := regexp.MustCompile(`تحويل مبلغ\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?([\d,]+\.\d{2}).*?جهة العمل`)
		matchSal := patternSal.FindStringSubmatch(body)

		if len(matchIPN) > 2 {
			tx.Currency = utils.NormalizeCurrency(matchIPN[1])
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchIPN[2], ",", ""), 64)
			tx.Amount = amount

			payeePattern := regexp.MustCompile(`from\s+(.*?)\s+with reference`)
			payeeMatch := payeePattern.FindStringSubmatch(body)
			if len(payeeMatch) > 1 {
				tx.Payee = strings.TrimSpace(payeeMatch[1])
			} else {
				tx.Payee = "Transfer In"
			}
		} else if len(matchSal) > 2 {
			tx.Currency = utils.NormalizeCurrency(matchSal[1])
			amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchSal[2], ",", ""), 64)
			tx.Amount = amount
			tx.Payee = "Salary / Work"
		}
	}
}
