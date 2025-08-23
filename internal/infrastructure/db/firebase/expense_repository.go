package firebase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"errors"

	"cloud.google.com/go/firestore"
	"github.com/theHinneh/budgeting/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExpenseRepository struct {
	Firestore *firestore.Client
}

func (f *ExpenseRepository) CreateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error) {
	if expense == nil || strings.TrimSpace(expense.UserID) == "" || strings.TrimSpace(expense.UID) == "" {
		return nil, fmt.Errorf("invalid expense")
	}
	_, err := f.Firestore.Collection("expenses").Doc(expense.UserID).Collection("expenses").Doc(expense.UID).Set(ctx, map[string]interface{}{
		"UID":                 expense.UID,
		"UserID":              expense.UserID,
		"Source":              expense.Source,
		"Amount":              expense.Amount,
		"Currency":            expense.Currency,
		"Notes":               expense.Notes,
		"IsRecurring":         expense.IsRecurring,
		"RecurrenceFrequency": expense.RecurrenceFrequency,
		"NextOccurrenceDate":  expense.NextOccurrenceDate,
		"CreatedAt":           expense.CreatedAt,
		"UpdatedAt":           expense.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return expense, nil
}

func (f *ExpenseRepository) ListExpensesByUser(ctx context.Context, userID string) ([]*domain.Expense, error) {
	var res []*domain.Expense
	iter := f.Firestore.Collection("expenses").Doc(userID).Collection("expenses").OrderBy("CreatedAt", firestore.Desc).Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		var m domain.Expense
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (f *ExpenseRepository) GetExpense(ctx context.Context, userID string, expenseID string) (*domain.Expense, error) {
	dsnap, err := f.Firestore.Collection("expenses").Doc(userID).Collection("expenses").Doc(expenseID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.Expense
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (f *ExpenseRepository) UpdateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error) {
	if expense == nil || strings.TrimSpace(expense.UserID) == "" || strings.TrimSpace(expense.UID) == "" {
		return nil, fmt.Errorf("invalid expense")
	}
	_, err := f.Firestore.Collection("expenses").Doc(expense.UserID).Collection("expenses").Doc(expense.UID).Set(ctx, map[string]interface{}{
		"UID":                 expense.UID,
		"UserID":              expense.UserID,
		"Source":              expense.Source,
		"Amount":              expense.Amount,
		"Currency":            expense.Currency,
		"Notes":               expense.Notes,
		"IsRecurring":         expense.IsRecurring,
		"RecurrenceFrequency": expense.RecurrenceFrequency,
		"NextOccurrenceDate":  expense.NextOccurrenceDate,
		"CreatedAt":           expense.CreatedAt,
		"UpdatedAt":           expense.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return expense, nil
}

func (f *ExpenseRepository) DeleteExpense(ctx context.Context, userID string, expenseID string) error {
	docRef := f.Firestore.Collection("expenses").Doc(userID).Collection("expenses").Doc(expenseID)
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Errorf(codes.NotFound, "expense not found")
		}
		return err
	}

	_, err = docRef.Delete(ctx)
	return err
}

func (f *ExpenseRepository) ListRecurringExpenses(ctx context.Context, userID string, before time.Time) ([]*domain.Expense, error) {
	var res []*domain.Expense
	iter := f.Firestore.Collection("expenses").Doc(userID).Collection("expenses").Where("IsRecurring", "==", true).Where("NextOccurrenceDate", "<=", before).Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		var m domain.Expense
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (f *ExpenseRepository) UpdateExpenseRecurringStatus(ctx context.Context, userID string, expenseID string, nextOccurrenceDate time.Time) error {
	_, err := f.Firestore.Collection("expenses").Doc(userID).Collection("expenses").Doc(expenseID).Update(ctx, []firestore.Update{
		{Path: "NextOccurrenceDate", Value: nextOccurrenceDate},
		{Path: "UpdatedAt", Value: time.Now().UTC()},
	})
	return err
}
