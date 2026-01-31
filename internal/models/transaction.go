package models

import "encoding/xml"

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
	CatGeneral   = "General"
)

// Transaction represents a parsed bank transaction
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

// TransactionType constants
const (
	TypeExpense = "Expense"
	TypeIncome  = "Income"
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
