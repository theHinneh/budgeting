package firebase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/theHinneh/budgeting/internal/core/models"
	"github.com/theHinneh/budgeting/internal/core/ports"
	"google.golang.org/api/iterator"
)

var _ ports.IncomeRepoPort = (*Database)(nil)

func (d *Database) CreateIncome(ctx context.Context, income *models.Income) (*models.Income, error) {
	if income == nil || income.UserID == "" || income.UID == "" {
		return nil, fmt.Errorf("invalid income")
	}
	_, err := d.Firestore.Collection("incomes").Doc(income.UserID).Collection("incomes").Doc(income.UID).Set(ctx, map[string]interface{}{
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

func (d *Database) ListIncomesByUser(ctx context.Context, userID string) ([]*models.Income, error) {
	var res []*models.Income
	iter := d.Firestore.Collection("incomes").Doc(userID).Collection("incomes").OrderBy("ReceivedAt", firestore.Desc).Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		var m models.Income
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (d *Database) GetIncome(ctx context.Context, userID string, incomeID string) (*models.Income, error) {
	dsnap, err := d.Firestore.Collection("incomes").Doc(userID).Collection("incomes").Doc(incomeID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m models.Income
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *Database) DeleteIncome(ctx context.Context, userID string, incomeID string) error {
	_, err := d.Firestore.Collection("incomes").Doc(userID).Collection("incomes").Doc(incomeID).Delete(ctx)
	return err
}

func (d *Database) CreateIncomeSource(ctx context.Context, src *models.IncomeSource) (*models.IncomeSource, error) {
	if src == nil || src.UserID == "" || src.UID == "" {
		return nil, fmt.Errorf("invalid income source")
	}
	_, err := d.Firestore.Collection("incomes").Doc(src.UserID).Collection("income_sources").Doc(src.UID).Set(ctx, map[string]interface{}{
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

func (d *Database) ListIncomeSourcesByUser(ctx context.Context, userID string) ([]*models.IncomeSource, error) {
	var res []*models.IncomeSource
	iter := d.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").OrderBy("Source", firestore.Asc).Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		var m models.IncomeSource
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (d *Database) ListDueIncomeSources(ctx context.Context, userID string, before time.Time) ([]*models.IncomeSource, error) {
	var res []*models.IncomeSource
	q := d.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").Where("Active", "==", true).Where("NextPayAt", "<=", before)
	iter := q.Documents(ctx)
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		var m models.IncomeSource
		if err := dsnap.DataTo(&m); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}
	return res, nil
}

func (d *Database) UpdateIncomeSource(ctx context.Context, userID string, id string, updates map[string]interface{}) error {
	var ups []firestore.Update
	for k, v := range updates {
		ups = append(ups, firestore.Update{Path: k, Value: v})
	}
	_, err := d.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").Doc(id).Update(ctx, ups)
	return err
}

func (d *Database) DeleteIncomeSource(ctx context.Context, userID string, source string) error {
	q := d.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").Where("Source", "==", source)
	iter := q.Documents(ctx)
	batch := d.Firestore.Batch()
	count := 0
	for {
		dsnap, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
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
