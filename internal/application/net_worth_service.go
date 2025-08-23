package application

import (
	"context"
	"strings"

	"github.com/theHinneh/budgeting/internal/application/ports"
)

type NetWorthService struct {
	incomeRepo  ports.IncomeRepoPort
	expenseRepo ports.ExpenseRepoPort
}

func NewNetWorthService(incomeRepo ports.IncomeRepoPort, expenseRepo ports.ExpenseRepoPort) *NetWorthService {
	return &NetWorthService{incomeRepo: incomeRepo, expenseRepo: expenseRepo}
}

var _ ports.NetWorthServicePort = (*NetWorthService)(nil)

func (s *NetWorthService) GetNetWorth(ctx context.Context, userID string) (*ports.NetWorthResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrValidation
	}

	incomes, err := s.incomeRepo.ListIncomesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	expenses, err := s.expenseRepo.ListExpensesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var totalIncome float64
	for _, income := range incomes {
		totalIncome += income.Amount
	}

	var totalExpense float64
	for _, expense := range expenses {
		totalExpense += expense.Amount
	}

	netWorth := totalIncome - totalExpense

	// Assuming a single currency for simplicity for now. A more robust solution
	// would handle multiple currencies or require a base currency for conversion.
	currency := "USD"
	if len(incomes) > 0 {
		currency = incomes[0].Currency
	} else if len(expenses) > 0 {
		currency = expenses[0].Currency
	}

	return &ports.NetWorthResponse{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		NetWorth:     netWorth,
		Currency:     currency,
	}, nil
}
