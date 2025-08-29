package dto

type NetWorthResponse struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	NetWorth     float64 `json:"net_worth"`
	Currency     string  `json:"currency"`
}
