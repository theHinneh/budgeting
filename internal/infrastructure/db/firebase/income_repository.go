package firebase

import (
	"context"
	"fmt"
	"time"

	"errors"

	"cloud.google.com/go/firestore"
	"github.com/theHinneh/budgeting/internal/domain"
	"google.golang.org/api/iterator"
)

type IncomeRepository struct {
	Firestore *firestore.Client
}

func (f *IncomeRepository) CreateIncome(ctx context.Context, income *domain.Income) (*domain.Income, error) {
	if income == nil || income.UserID == "" || income.UID == "" {
		return nil, fmt.Errorf("invalid income")
	}
	_, err := f.Firestore.Collection("incomes").Doc(income.UserID).Collection("incomes").Doc(income.UID).Set(ctx, map[string]interface{}{
		"UID":       income.UID,
		"UserID":    income.UserID,
		"Source":    income.Source,
		"Amount":    income.Amount,
		"Currency":  income.Currency,
		"Notes":     income.Notes,
		"CreatedAt": income.CreatedAt,
		"UpdatedAt": income.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return income, nil
}

func (f *IncomeRepository) ListIncomesByUser(ctx context.Context, userID string) ([]*domain.Income, error) {
	var res []*domain.Income
	iter := f.Firestore.Collection("incomes").Doc(userID).Collection("incomes").OrderBy("CreatedAt", firestore.Desc).Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		var m domain.Income
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (f *IncomeRepository) GetIncome(ctx context.Context, userID string, incomeID string) (*domain.Income, error) {
	dsnap, err := f.Firestore.Collection("incomes").Doc(userID).Collection("incomes").Doc(incomeID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.Income
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (f *IncomeRepository) DeleteIncome(ctx context.Context, userID string, incomeID string) error {
	_, err := f.Firestore.Collection("incomes").Doc(userID).Collection("incomes").Doc(incomeID).Delete(ctx)
	return err
}

func (f *IncomeRepository) CreateIncomeSource(ctx context.Context, src *domain.IncomeSource) (*domain.IncomeSource, error) {
	if src == nil || src.UserID == "" || src.UID == "" {
		return nil, fmt.Errorf("invalid income source")
	}
	_, err := f.Firestore.Collection("incomes").Doc(src.UserID).Collection("income_sources").Doc(src.UID).Set(ctx, map[string]interface{}{
		"UID":       src.UID,
		"UserID":    src.UserID,
		"Source":    src.Source,
		"Amount":    src.Amount,
		"Currency":  src.Currency,
		"Frequency": src.Frequency,
		"NextPayAt": src.NextPayAt,
		"Active":    src.Active,
		"Notes":     src.Notes,
		"CreatedAt": src.CreatedAt,
		"UpdatedAt": src.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return src, nil
}

func (f *IncomeRepository) ListIncomeSourcesByUser(ctx context.Context, userID string) ([]*domain.IncomeSource, error) {
	var res []*domain.IncomeSource
	iter := f.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").OrderBy("Source", firestore.Asc).Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		var m domain.IncomeSource
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (f *IncomeRepository) ListDueIncomeSources(ctx context.Context, userID string, before time.Time) ([]*domain.IncomeSource, error) {
	var res []*domain.IncomeSource
	q := f.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").Where("Active", "==", true).Where("NextPayAt", "<=", before)
	iter := q.Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		var m domain.IncomeSource
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (f *IncomeRepository) UpdateIncomeSource(ctx context.Context, userID string, id string, updates map[string]interface{}) error {
	var ups []firestore.Update
	for k, v := range updates {
		ups = append(ups, firestore.Update{Path: k, Value: v})
	}
	_, err := f.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").Doc(id).Update(ctx, ups)
	return err
}

func (f *IncomeRepository) DeleteIncomeSource(ctx context.Context, userID string, source string) error {
	q := f.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").Where("Source", "==", source)
	iter := q.Documents(ctx)
	batch := f.Firestore.Batch()
	count := 0
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return err
		}
		batch.Delete(dsnap.Ref)
		count++
	}
	if count == 0 {
		return nil
	}
	_, err := batch.Commit(ctx)
	return err
}
