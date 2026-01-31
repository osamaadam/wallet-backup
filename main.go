package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Category constants
const (
	CatFood      = "Food & Drink"
	CatShopping  = "Shopping"
	CatHousing   = "Housing"
	CatTransport = "Transportation"
	CatVehicle   = "Vehicle"
	CatLife      = "Life & Entertainment"
	CatComms     = "Communication, PC"
	CatFinancial = "Financial expenses"
	CatIncome    = "Income"
)

// SMS represents a single SMS message from the XML backup
type SMS struct {
	Address string `xml:"address,attr"`
	Body    string `xml:"body,attr"`
	Date    string `xml:"date,attr"`
}

// SMSBackup represents the root of the XML document
type SMSBackup struct {
	XMLName xml.Name `xml:"smses"`
	SMS     []SMS    `xml:"sms"`
}

// Transaction represents a parsed transaction
type Transaction struct {
	Date        string
	Payee       string
	Amount      float64
	Currency    string
	Type        string
	Category    string
	Note        string
	TargetGroup string
}

var (
	outputDir string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "sms-parser [xml-file]",
		Short: "Parse SMS backup and extract bank transactions",
		Long:  `A CLI tool to parse SMS backup XML files and extract bank transactions into CSV files.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runParser,
	}

	rootCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for CSV files (created if not exists)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runParser(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parse the SMS backup file
	transactions, err := parseSMSBackup(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse SMS backup: %w", err)
	}

	// Write transactions to CSV files
	if err := writeTransactions(transactions); err != nil {
		return fmt.Errorf("failed to write transactions: %w", err)
	}

	return nil
}

func parseSMSBackup(filePath string) (map[string][]Transaction, error) {
	// Read and parse XML file
	xmlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var backup SMSBackup
	if err := xml.Unmarshal(xmlFile, &backup); err != nil {
		return nil, fmt.Errorf("error parsing XML: %w", err)
	}

	groupedData := map[string][]Transaction{
		"CIB_Current_Debit": {},
		"Banque_Misr":       {},
	}

	seenTransactions := make(map[string]bool)

	for _, sms := range backup.SMS {
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
		dateStr := dateObj.Format("2006-01-02 15:04:05")

		tx := Transaction{
			Date:     dateStr,
			Payee:    "",
			Amount:   0.0,
			Currency: "EGP",
			Type:     "Expense",
			Category: "General",
			Note:     sms.Body,
		}

		// Parse based on sender
		if sms.Address == "CIB" {
			parseCIBMessage(&tx, sms.Body)
		} else if sms.Address == "Banque Misr" {
			parseBanqueMisrMessage(&tx, sms.Body)
		}

		// Apply categorization
		if tx.TargetGroup != "" && tx.Amount != 0 && tx.Category == "General" {
			tx.Category = getCategory(tx.Payee, tx.Note, tx.Amount)
		}

		// Add category to note and append to group
		if tx.TargetGroup != "" && tx.Amount != 0 {
			if _, exists := groupedData[tx.TargetGroup]; !exists {
				groupedData[tx.TargetGroup] = []Transaction{}
			}

			if tx.Category != "General" {
				tx.Note = fmt.Sprintf("[%s] %s", tx.Category, tx.Note)
			}

			groupedData[tx.TargetGroup] = append(groupedData[tx.TargetGroup], tx)
		}
	}

	return groupedData, nil
}

func parseCIBMessage(tx *Transaction, body string) {
	// Detect credit card
	ccPattern := regexp.MustCompile(`(?i)(?:credit card|ending with|card|بـ)\s*[#*]*\s*(\d{4})`)
	ccMatch := ccPattern.FindStringSubmatch(body)

	isCreditCard := false
	cardDigits := "Unknown"

	if len(ccMatch) > 1 {
		cardDigits = ccMatch[1]
		if cardDigits != "7759" && cardDigits != "2373" {
			isCreditCard = true
			tx.TargetGroup = fmt.Sprintf("CIB_Credit_Card_%s", cardDigits)
		}
	}

	if isCreditCard {
		// Credit card transactions
		if strings.Contains(body, "charged for") || strings.Contains(body, "purchasing transaction") {
			pattern := regexp.MustCompile(`(?i)charged for\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})\s*at\s*(.*?)(?:\s+on|\s+at|\. Available)`)
			match := pattern.FindStringSubmatch(body)
			if len(match) > 3 {
				tx.Currency = normalizeCurrency(match[1])
				amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
				tx.Amount = -amount
				tx.Payee = cleanPayeeName(strings.TrimSpace(match[3]))
			}
		} else if strings.Contains(body, "refunded") || strings.Contains(body, "rad") || strings.Contains(body, "رد") {
			if !strings.Contains(body, "تم سداد") {
				tx.Type = "Income"
				pattern := regexp.MustCompile(`(?i)(?:refunded|red|rd|رد)\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
				match := pattern.FindStringSubmatch(body)
				if len(match) > 2 {
					tx.Currency = normalizeCurrency(match[1])
					amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
					tx.Amount = amount
					tx.Payee = "Refund"
				}
			}
		}

		if strings.Contains(body, "تم سداد") || (strings.Contains(body, "payment") && strings.Contains(body, "received")) {
			tx.Type = "Income"
			tx.Payee = "CIB Repayment"
			pattern := regexp.MustCompile(`مبلغ\s*([\d,]+\.\d{2})`)
			match := pattern.FindStringSubmatch(body)
			if len(match) > 1 {
				amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
				tx.Amount = amount
			}
		}
	} else if strings.Contains(body, "7759") || strings.Contains(body, "2373") {
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
				tx.Currency = normalizeCurrency(matchAr[1])
				amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchAr[2], ",", ""), 64)
				tx.Amount = -amount
				tx.Payee = cleanPayeeName(strings.TrimSpace(matchAr[3]))
			} else if len(matchEn) > 3 {
				tx.Currency = normalizeCurrency(matchEn[1])
				amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchEn[2], ",", ""), 64)
				tx.Amount = -amount
				tx.Payee = cleanPayeeName(strings.TrimSpace(matchEn[3]))
			} else if len(matchWith) > 2 {
				tx.Currency = normalizeCurrency(matchWith[1])
				amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchWith[2], ",", ""), 64)
				tx.Amount = -amount
				tx.Payee = "ATM Withdrawal"
			}
		} else if strings.Contains(body, "2373") {
			if strings.Contains(body, "debited") || strings.Contains(body, "charged with") || strings.Contains(body, "تم تحويل") {
				pattern := regexp.MustCompile(`(?i)(?:amount|for)\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
				match := pattern.FindStringSubmatch(body)
				if len(match) > 2 {
					tx.Currency = normalizeCurrency(match[1])
					amount, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
					tx.Amount = -amount

					if strings.Contains(body, "transfer to another account") {
						tx.Payee = "Transfer to Account / CC"
						tx.Category = CatFinancial
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
				tx.Type = "Income"

				// IPN pattern
				patternIPN := regexp.MustCompile(`(?i)credited with IPN Inward for\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
				matchIPN := patternIPN.FindStringSubmatch(body)

				// Salary pattern
				patternSal := regexp.MustCompile(`تحويل مبلغ\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?([\d,]+\.\d{2}).*?جهة العمل`)
				matchSal := patternSal.FindStringSubmatch(body)

				if len(matchIPN) > 2 {
					tx.Currency = normalizeCurrency(matchIPN[1])
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
					tx.Currency = normalizeCurrency(matchSal[1])
					amount, _ := strconv.ParseFloat(strings.ReplaceAll(matchSal[2], ",", ""), 64)
					tx.Amount = amount
					tx.Payee = "Salary / Work"
				}
			}
		}
	}
}

func parseBanqueMisrMessage(tx *Transaction, body string) {
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
		pattern := regexp.MustCompile(`مبلغ\s*(?:([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)\s*)?([\d,]+)(?:\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم))?`)
		match := pattern.FindStringSubmatch(body)

		if len(match) > 2 {
			val, _ := strconv.ParseFloat(strings.ReplaceAll(match[2], ",", ""), 64)
			detectedCurr := match[1]
			if detectedCurr == "" {
				detectedCurr = match[3]
			}
			tx.Currency = normalizeCurrency(detectedCurr)

			if strings.Contains(body, "من حساب") {
				tx.Amount = -val
				tx.Payee = "Transfer Out"
			} else if strings.Contains(body, "الى حساب") {
				tx.Type = "Income"
				tx.Amount = val
				tx.Payee = "Transfer In"
			}
		}
	} else if strings.Contains(body, "تم الخصم") || strings.Contains(body, "transaction") {
		pattern := regexp.MustCompile(`(?:مبلغ|amount)\s*([A-Za-z]{3}|L\.E\.?|ج\.م|جنيه|جم)?\s*([\d,]+\.\d{2})`)
		match := pattern.FindStringSubmatch(body)

		if len(match) > 2 {
			tx.Currency = normalizeCurrency(match[1])
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
}

func normalizeCurrency(currStr string) string {
	if currStr == "" {
		return "EGP"
	}

	cleanCurr := strings.ToUpper(strings.TrimSpace(currStr))
	mapping := map[string]string{
		"LE":   "EGP",
		"L.E":  "EGP",
		"L.E.": "EGP",
		"EGP":  "EGP",
		"ج.م":  "EGP",
		"جم":   "EGP",
		"جنيه": "EGP",
		"USD":  "USD",
		"EUR":  "EUR",
		"GBP":  "GBP",
		"TRY":  "TRY",
		"JPY":  "JPY",
	}

	if normalized, ok := mapping[cleanCurr]; ok {
		return normalized
	}
	return cleanCurr
}

func cleanPayeeName(payeeRaw string) string {
	if payeeRaw == "" {
		return ""
	}

	prefixes := []string{
		"PAYMOB-", "PAYMOB ", "PAYMOBS ", "GEIDEA ", "GEIDEAE ",
		"FAWRY ", "FAWRYPF ", "MY FAWRY", "Fawry ", "FawryPF ",
		"AFS-", "AFS ", "POS ", "NGOV_UNI ", "BEE ", "KASHIER ",
	}

	clean := payeeRaw
	for _, p := range prefixes {
		if strings.HasPrefix(strings.ToUpper(clean), strings.ToUpper(p)) {
			clean = strings.TrimSpace(clean[len(p):])
		}
	}

	// Remove trailing digits
	digitsPattern := regexp.MustCompile(`\s*\d+$`)
	clean = digitsPattern.ReplaceAllString(clean, "")

	return strings.TrimSpace(clean)
}

func getCategory(payee, note string, amount float64) string {
	cleanPayee := cleanPayeeName(payee)
	text := strings.ToLower(cleanPayee + " " + note)

	// Income
	if amount > 0 {
		return CatIncome
	}

	// Financial / Transfers
	if contains(text, "credit card payment", "sadaad", "cib repayment") {
		return CatFinancial
	}

	// Shopping
	shoppingKeywords := []string{
		"amazon", "noon", "jumia", "souq", "shopping", "zara", "h&m",
		"lc waikiki", "defacto", "american eagle", "lachica", "ravin",
		"el salama", "stitch", "clothes", "fashion", "shoes", "concrete",
		"town team", "activ", "naga", "rich for cloth", "pronto",
		"scarpe", "scarape", "tie house", "rose paris", "b tech", "b.tech",
		"trade line", "2b", "best buy", "dubai phone", "mobile shop",
		"el araby", "fresh electric", "tornado",
	}
	if containsAny(text, shoppingKeywords...) {
		return CatShopping
	}

	// Housing (furniture)
	if containsAny(text, "ikea", "homzmart", "furniture", "jotun", "ahfad") {
		return CatHousing
	}

	// Food & Drink
	foodKeywords := []string{
		"mcdonalds", "kfc", "pizza", "burger", "buffalo", "primos",
		"spectra", "desoky", "sandwich", "elmenus", "talabat", "breadfast",
		"roosters", "hardees", "manchow", "willys", "dhad", "el dahan",
		"sanabel", "fookotcharia", "krispy", "cafe", "costa", "starbucks",
		"cilantro", "tbsp", "espresso", "beano", "cinnabon", "dunkin",
		"caribou", "house of cocoa", "sale sucre", "dar el bon", "karak",
		"potasta", "b labn", "b.labn", "carrefour", "fathalla", "market",
		"seoudi", "gomla", "bim", "kazyon", "hyper", "ramadan hamada",
		"saood", "metro", "kheir zaman", "ragab", "abu auf", "kashier",
		"elkhalil", "aswak", "fresh food", "sun mall", "grapes",
	}
	if containsAny(text, foodKeywords...) {
		return CatFood
	}

	// Transportation
	transportKeywords := []string{
		"uber", "didi", "careem", "indriver", "transport", "super jet",
		"railways", "go bus", "swvl", "pegasus", "fly", "airline",
		"booking", "flight",
	}
	if containsAny(text, transportKeywords...) {
		return CatTransport
	}

	// Vehicle
	vehicleKeywords := []string{
		"mobil", "chillout", "gas station", "total", "ola", "master gas",
		"adnoc", "wataniya", "fuel", "car service", "tire", "fit & fix",
	}
	if containsAny(text, vehicleKeywords...) {
		return CatVehicle
	}

	// Housing & Utilities
	housingKeywords := []string{
		"sahl", "electricity", "water", "bill", "national gas", "natgas",
		"town gas", "petrotrade", "taqa", "north cairo",
	}
	if containsAny(text, housingKeywords...) {
		return CatHousing
	}

	// Communication & PC
	commsKeywords := []string{
		"vodafone", "orange", "etisalat", "we ", "telecom", "top up",
		"landline", "we-fv", "internet", "fbb", "adsl", "google",
		"microsoft", "adobe", "apple", "icloud", "storage", "host",
		"domain", "xbox", "playstation", "steam", "games", "mullvad",
		"linkedin",
	}
	if containsAny(text, commsKeywords...) {
		return CatComms
	}

	// Life & Entertainment
	lifeKeywords := []string{
		"netflix", "spotify", "osn", "shahid", "youtube", "watch it",
		"yango", "vox", "cinema", "renessance", "ticket", "tazkarti",
		"kindle", "audible", "books", "diwan", "pharmacy", "dr.",
		"hospital", "medical", "ezaby", "elezzaby", "seif", "rushdy",
		"andalusia", "yosra", "hany", "tay",
	}
	if containsAny(text, lifeKeywords...) {
		return CatLife
	}

	// Financial / Cash
	financialKeywords := []string{
		"atm", "withdrawal", "s7b", "سحب", "cash", "fawry",
		"my fawry", "fawrypay",
	}
	if containsAny(text, financialKeywords...) {
		return CatFinancial
	}

	return "General"
}

func contains(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func containsAny(text string, keywords ...string) bool {
	return contains(text, keywords...)
}

func writeTransactions(groupedData map[string][]Transaction) error {
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
		filename := filepath.Join(outputDir, groupName+".csv")
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("error creating %s: %w", filename, err)
		}
		defer file.Close()

		// Write BOM for UTF-8
		file.Write([]byte{0xEF, 0xBB, 0xBF})

		writer := csv.NewWriter(file)
		writer.Comma = ';'

		// Write header
		if err := writer.Write(fieldnames); err != nil {
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

		fmt.Printf("Created %s with %d transactions.\n", filename, len(transactions))
	}

	return nil
}
