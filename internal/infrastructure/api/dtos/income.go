package dtos

import (
	"time"

	"github.com/theHinneh/budgeting/internal/domain"
)

type AddIncomeRequest struct {
	Source   string  `json:"source" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency,omitempty"`
	Notes    string  `json:"notes,omitempty"`
}

type IncomeResponse struct {
	UID       string    `json:"uid"`
	UserID    string    `json:"user_id"`
	Source    string    `json:"source"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListIncomeResponse struct {
	Incomes []*IncomeResponse `json:"incomes"`
	Count   int               `json:"count"`
}

func NewIncomeResponse(income *domain.Income) *IncomeResponse {
	if income == nil {
		return nil
	}
	return &IncomeResponse{
		UID:       income.UID,
		UserID:    income.UserID,
		Source:    income.Source,
		Amount:    income.Amount,
		Currency:  income.Currency,
		Notes:     income.Notes,
		CreatedAt: income.CreatedAt,
		UpdatedAt: income.UpdatedAt,
	}
}

func NewListIncomeResponse(incomes []*domain.Income) *ListIncomeResponse {
	resps := make([]*IncomeResponse, len(incomes))
	for i, income := range incomes {
		resps[i] = NewIncomeResponse(income)
	}
	return &ListIncomeResponse{
		Incomes: resps,
		Count:   len(resps),
	}
}

type AddIncomeSourceRequest struct {
	Source    string  `json:"source" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Currency  string  `json:"currency,omitempty"`
	Frequency string  `json:"frequency" binding:"required,oneof=weekly biweekly monthly"`
	NextPayAt string  `json:"next_pay_at,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

type IncomeSourceResponse struct {
	UID       string    `json:"uid"`
	UserID    string    `json:"user_id"`
	Source    string    `json:"source"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency,omitempty"`
	Frequency string    `json:"frequency"`
	NextPayAt time.Time `json:"next_pay_at"`
	Active    bool      `json:"active"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListIncomeSourceResponse struct {
	Sources []*IncomeSourceResponse `json:"sources"`
	Count   int                     `json:"count"`
}

func NewIncomeSourceResponse(source *domain.IncomeSource) *IncomeSourceResponse {
	if source == nil {
		return nil
	}
	return &IncomeSourceResponse{
		UID:       source.UID,
		UserID:    source.UserID,
		Source:    source.Source,
		Amount:    source.Amount,
		Currency:  source.Currency,
		Frequency: source.Frequency,
		NextPayAt: source.NextPayAt,
		Active:    source.Active,
		Notes:     source.Notes,
		CreatedAt: source.CreatedAt,
		UpdatedAt: source.UpdatedAt,
	}
}

func NewListIncomeSourceResponse(sources []*domain.IncomeSource) *ListIncomeSourceResponse {
	resps := make([]*IncomeSourceResponse, len(sources))
	for i, source := range sources {
		resps[i] = NewIncomeSourceResponse(source)
	}
	return &ListIncomeSourceResponse{
		Sources: resps,
		Count:   len(resps),
	}
}
