package utils

import (
	"regexp"
	"strings"
)

// NormalizeCurrency converts various currency representations to standard codes
func NormalizeCurrency(currStr string) string {
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

// CleanPayeeName removes payment processor prefixes and trailing digits
func CleanPayeeName(payeeRaw string) string {
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
			break
		}
	}

	// Remove trailing digits
	digitsPattern := regexp.MustCompile(`\s*\d+$`)
	clean = digitsPattern.ReplaceAllString(clean, "")

	return strings.TrimSpace(clean)
}

// Contains checks if text contains any of the given keywords
func Contains(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
