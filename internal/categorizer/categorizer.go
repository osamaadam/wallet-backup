package categorizer

import (
	"strings"

	"sms-parser/internal/models"
	"sms-parser/internal/utils"
)

// Categorizer handles transaction categorization
type Categorizer struct{}

// New creates a new Categorizer instance
func New() *Categorizer {
	return &Categorizer{}
}

// Categorize assigns a category to a transaction based on payee and note
func (c *Categorizer) Categorize(payee, note string, amount float64) string {
	cleanPayee := utils.CleanPayeeName(payee)
	text := strings.ToLower(cleanPayee + " " + note)

	// Income
	if amount > 0 {
		return models.CatIncome
	}

	// Financial / Transfers
	if utils.Contains(text, "credit card payment", "sadaad", "cib repayment") {
		return models.CatFinancial
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
	if utils.Contains(text, shoppingKeywords...) {
		return models.CatShopping
	}

	// Housing (furniture)
	if utils.Contains(text, "ikea", "homzmart", "furniture", "jotun", "ahfad") {
		return models.CatHousing
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
	if utils.Contains(text, foodKeywords...) {
		return models.CatFood
	}

	// Transportation
	transportKeywords := []string{
		"uber", "didi", "careem", "indriver", "transport", "super jet",
		"railways", "go bus", "swvl", "pegasus", "fly", "airline",
		"booking", "flight",
	}
	if utils.Contains(text, transportKeywords...) {
		return models.CatTransport
	}

	// Vehicle
	vehicleKeywords := []string{
		"mobil", "chillout", "gas station", "total", "ola", "master gas",
		"adnoc", "wataniya", "fuel", "car service", "tire", "fit & fix",
	}
	if utils.Contains(text, vehicleKeywords...) {
		return models.CatVehicle
	}

	// Housing & Utilities
	housingKeywords := []string{
		"sahl", "electricity", "water", "bill", "national gas", "natgas",
		"town gas", "petrotrade", "taqa", "north cairo",
	}
	if utils.Contains(text, housingKeywords...) {
		return models.CatHousing
	}

	// Communication & PC
	commsKeywords := []string{
		"vodafone", "orange", "etisalat", "we ", "telecom", "top up",
		"landline", "we-fv", "internet", "fbb", "adsl", "google",
		"microsoft", "adobe", "apple", "icloud", "storage", "host",
		"domain", "xbox", "playstation", "steam", "games", "mullvad",
		"linkedin",
	}
	if utils.Contains(text, commsKeywords...) {
		return models.CatComms
	}

	// Life & Entertainment
	lifeKeywords := []string{
		"netflix", "spotify", "osn", "shahid", "youtube", "watch it",
		"yango", "vox", "cinema", "renessance", "ticket", "tazkarti",
		"kindle", "audible", "books", "diwan", "pharmacy", "dr.",
		"hospital", "medical", "ezaby", "elezzaby", "seif", "rushdy",
		"andalusia", "yosra", "hany", "tay",
	}
	if utils.Contains(text, lifeKeywords...) {
		return models.CatLife
	}

	// Financial / Cash
	financialKeywords := []string{
		"atm", "withdrawal", "s7b", "سحب", "cash", "fawry",
		"my fawry", "fawrypay",
	}
	if utils.Contains(text, financialKeywords...) {
		return models.CatFinancial
	}

	return models.CatGeneral
}
