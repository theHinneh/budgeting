package dto

type AddIncomeInput struct {
	UserID   string
	Source   string
	Amount   float64
	Currency string
	Notes    string
}

type AddIncomeSourceInput struct {
	UserID    string
	Source    string
	Amount    float64
	Currency  string
	Frequency PayFrequency
	NextPayAt string
	Notes     string
}

type IncomeUpdateInput struct {
	IncomeID      string
	UserID        string
	PreviousValue float64
	CurrentValue  float64
}

type PayFrequency string

const (
	PayWeekly   PayFrequency = "weekly"
	PayBiWeekly PayFrequency = "biweekly"
	PayMonthly  PayFrequency = "monthly"
)
