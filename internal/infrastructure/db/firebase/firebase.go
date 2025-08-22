package firebase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"errors"

	"cloud.google.com/go/firestore"
	fb "firebase.google.com/go/v4"
	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/domain"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Database struct {
	app       *fb.App
	Firestore *firestore.Client
	Auth      *fbAuth.Client
}

func NewDatabase(ctx context.Context, cfg *config.Configuration) (*Database, error) {
	getStr := func(primary string, fallbacks ...string) string {
		if v := cfg.V.GetString(primary); v != "" {
			return v
		}
		for _, fbk := range fallbacks {
			if v := cfg.V.GetString(fbk); v != "" {
				return v
			}
		}
		return ""
	}

	projectID := getStr("FIREBASE_PROJECT_ID", "database.FIREBASE_PROJECT_ID", "database.firebase_project_id")
	credsFile := getStr("FIREBASE_CREDENTIALS_FILE", "database.FIREBASE_CREDENTIALS_FILE", "database.firebase_credentials_file")
	srvsAccID := getStr("FIREBASE_SA_ID", "database.FIREBASE_SA_ID", "database.firebase_sa_id")

	var opts []option.ClientOption
	if credsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credsFile))
	}

	fbCfg := &fb.Config{}
	if projectID != "" {
		fbCfg.ProjectID = projectID
		fbCfg.ServiceAccountID = srvsAccID
	}

	app, err := fb.NewApp(ctx, fbCfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase app init failed: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase auth init failed: %w", err)
	}

	fsClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore init failed: %w", err)
	}

	return &Database{app: app, Firestore: fsClient, Auth: authClient}, nil
}

func (d *Database) AutoMigrate(models ...interface{}) error { return nil }

func (d *Database) Close() error {
	if d.Firestore != nil {
		return d.Firestore.Close()
	}
	return nil
}

func (d *Database) Ping(ctx context.Context) error {
	if d.Firestore == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	iter := d.Firestore.Collections(ctx)
	_, err := iter.Next()
	if err != nil && err != iterator.Done {
		// Any error other than iterator.Done indicates a connectivity/config problem.
		return err
	}

	return nil
}

var _ ports.UserPersistencePort = (*Database)(nil)
var _ ports.IncomeRepoPort = (*Database)(nil)
var _ ports.ExpenseRepoPort = (*Database)(nil)

// --- Methods from UserPersistencePort (Auth-related / Profile Persistence) ---

func (d *Database) CreateAuthUser(ctx context.Context, email, password, displayName string, phone *string) (string, error) {
	params := (&fbAuth.UserToCreate{}).
		Email(email).
		Password(password).
		DisplayName(displayName)
	if phone != nil && strings.TrimSpace(*phone) != "" {
		params = params.PhoneNumber(*phone)
	}
	u, err := d.Auth.CreateUser(ctx, params)
	if err != nil {
		return "", err
	}
	return u.UID, nil
}

func (d *Database) GetAuthUser(ctx context.Context, uid string) error {
	_, err := d.Auth.GetUser(ctx, uid)
	return err
}

func (d *Database) UpdateAuthUser(ctx context.Context, uid string, email *string, displayName *string, phone *string) error {
	var upd *fbAuth.UserToUpdate
	if email != nil {
		if upd == nil {
			upd = &fbAuth.UserToUpdate{}
		}
		upd = upd.Email(*email)
	}
	if displayName != nil {
		if upd == nil {
			upd = &fbAuth.UserToUpdate{}
		}
		upd = upd.DisplayName(*displayName)
	}
	if phone != nil {
		if upd == nil {
			upd = &fbAuth.UserToUpdate{}
		}
		if strings.TrimSpace(*phone) != "" {
			upd = upd.PhoneNumber(*phone)
		}
	}
	if upd == nil {
		return nil
	}
	_, err := d.Auth.UpdateUser(ctx, uid, upd)
	return err
}

func (d *Database) DeleteAuthUser(ctx context.Context, uid string) error {
	return d.Auth.DeleteUser(ctx, uid)
}

func (d *Database) UpdatePassword(ctx context.Context, uid string, newPassword string) error {
	upd := (&fbAuth.UserToUpdate{}).Password(newPassword)
	_, err := d.Auth.UpdateUser(ctx, uid, upd)
	return err
}

func (d *Database) GeneratePasswordResetLink(ctx context.Context, email string) (string, error) {
	return d.Auth.PasswordResetLink(ctx, email)
}

// CreateUser saves the user profile to Firestore
func (d *Database) CreateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	if u == nil || strings.TrimSpace(u.UID) == "" {
		return nil, fmt.Errorf("invalid user profile")
	}
	_, err := d.Firestore.Collection("users").Doc(u.UID).Set(ctx, map[string]interface{}{
		"UID":           u.UID,
		"Username":      u.Username,
		"Email":         u.Email,
		"FirstName":     u.FirstName,
		"LastName":      u.LastName,
		"PhoneNumber":   u.PhoneNumber,
		"ProviderID":    u.ProviderID,
		"PhotoURL":      u.PhotoURL,
		"EmailVerified": u.EmailVerified,
		"CreatedAt":     u.CreatedAt,
		"UpdatedAt":     u.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUser retrieves the user profile from Firestore
func (d *Database) GetUser(ctx context.Context, uid string) (*domain.User, error) {
	dsnap, err := d.Firestore.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.User
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *Database) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, fmt.Errorf("GetUserByEmail not directly supported/efficient for Firestore without a specific index. Implement a query on 'Email' field if needed.")
}

func (d *Database) UpdateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	if u == nil || strings.TrimSpace(u.UID) == "" {
		return nil, fmt.Errorf("invalid user profile for update")
	}

	updates := map[string]interface{}{
		"Username":      u.Username,
		"Email":         u.Email,
		"FirstName":     u.FirstName,
		"LastName":      u.LastName,
		"PhoneNumber":   u.PhoneNumber,
		"ProviderID":    u.ProviderID,
		"PhotoURL":      u.PhotoURL,
		"EmailVerified": u.EmailVerified,
		"UpdatedAt":     u.UpdatedAt,
	}

	var firestoreUpdates []firestore.Update
	for k, v := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: k, Value: v})
	}

	_, err := d.Firestore.Collection("users").Doc(u.UID).Update(ctx, firestoreUpdates)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (d *Database) DeleteUser(ctx context.Context, uid string) error {
	_, err := d.Firestore.Collection("users").Doc(uid).Delete(ctx)
	return err
}

func (d *Database) SaveProfile(ctx context.Context, u *domain.User) error {
	_, err := d.CreateUser(ctx, u)
	return err
}

func (d *Database) GetProfile(ctx context.Context, uid string) (*domain.User, error) {
	return d.GetUser(ctx, uid)
}

func (d *Database) UpdateProfile(ctx context.Context, uid string, updates map[string]interface{}) error {
	existingUser, err := d.GetUser(ctx, uid)
	if err != nil {
		return err
	}

	for key, value := range updates {
		switch key {
		case "Username":
			if v, ok := value.(string); ok {
				existingUser.Username = v
			}
		case "Email":
			if v, ok := value.(string); ok {
				existingUser.Email = v
			}
		case "FirstName":
			if v, ok := value.(string); ok {
				existingUser.FirstName = v
			}
		case "LastName":
			if v, ok := value.(string); ok {
				existingUser.LastName = v
			}
		case "PhoneNumber":
			if v, ok := value.(*string); ok {
				existingUser.PhoneNumber = v
			}
		case "ProviderID":
			if v, ok := value.(string); ok {
				existingUser.ProviderID = v
			}
		case "PhotoURL":
			if v, ok := value.(string); ok {
				existingUser.PhotoURL = v
			}
		case "EmailVerified":
			if v, ok := value.(bool); ok {
				existingUser.EmailVerified = v
			}
		case "UpdatedAt":
			if v, ok := value.(time.Time); ok {
				existingUser.UpdatedAt = v
			}
		}
	}
	_, err = d.UpdateUser(ctx, existingUser)
	return err
}

func (d *Database) DeleteProfile(ctx context.Context, uid string) error {
	return d.DeleteUser(ctx, uid)
}

// --- IncomeRepoPort Implementation ---

func (d *Database) CreateIncome(ctx context.Context, income *domain.Income) (*domain.Income, error) {
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

func (d *Database) ListIncomesByUser(ctx context.Context, userID string) ([]*domain.Income, error) {
	var res []*domain.Income
	iter := d.Firestore.Collection("incomes").Doc(userID).Collection("incomes").OrderBy("CreatedAt", firestore.Desc).Documents(ctx)
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

func (d *Database) GetIncome(ctx context.Context, userID string, incomeID string) (*domain.Income, error) {
	dsnap, err := d.Firestore.Collection("incomes").Doc(userID).Collection("incomes").Doc(incomeID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.Income
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *Database) DeleteIncome(ctx context.Context, userID string, incomeID string) error {
	_, err := d.Firestore.Collection("incomes").Doc(userID).Collection("incomes").Doc(incomeID).Delete(ctx)
	return err
}

func (d *Database) CreateIncomeSource(ctx context.Context, src *domain.IncomeSource) (*domain.IncomeSource, error) {
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

func (d *Database) ListIncomeSourcesByUser(ctx context.Context, userID string) ([]*domain.IncomeSource, error) {
	var res []*domain.IncomeSource
	iter := d.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").OrderBy("Source", firestore.Asc).Documents(ctx)
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

func (d *Database) ListDueIncomeSources(ctx context.Context, userID string, before time.Time) ([]*domain.IncomeSource, error) {
	var res []*domain.IncomeSource
	q := d.Firestore.Collection("incomes").Doc(userID).Collection("income_sources").Where("Active", "==", true).Where("NextPayAt", "<=", before)
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

// --- ExpenseRepoPort Implementation ---

func (d *Database) CreateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error) {
	if expense == nil || expense.UserID == "" || expense.UID == "" {
		return nil, fmt.Errorf("invalid expense")
	}
	_, err := d.Firestore.Collection("expenses").Doc(expense.UserID).Collection("expenses").Doc(expense.UID).Set(ctx, map[string]interface{}{
		"UID":       expense.UID,
		"UserID":    expense.UserID,
		"Source":    expense.Source,
		"Amount":    expense.Amount,
		"Currency":  expense.Currency,
		"Notes":     expense.Notes,
		"CreatedAt": expense.CreatedAt,
		"UpdatedAt": expense.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return expense, nil
}

func (d *Database) ListExpensesByUser(ctx context.Context, userID string) ([]*domain.Expense, error) {
	var res []*domain.Expense
	iter := d.Firestore.Collection("expenses").Doc(userID).Collection("expenses").OrderBy("CreatedAt", firestore.Desc).Documents(ctx)
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

func (d *Database) GetExpense(ctx context.Context, userID string, expenseID string) (*domain.Expense, error) {
	dsnap, err := d.Firestore.Collection("expenses").Doc(userID).Collection("expenses").Doc(expenseID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.Expense
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *Database) UpdateExpense(ctx context.Context, expense *domain.Expense) (*domain.Expense, error) {
	if expense == nil || expense.UserID == "" || expense.UID == "" {
		return nil, fmt.Errorf("invalid expense")
	}
	_, err := d.Firestore.Collection("expenses").Doc(expense.UserID).Collection("expenses").Doc(expense.UID).Set(ctx, map[string]interface{}{
		"UID":       expense.UID,
		"UserID":    expense.UserID,
		"Source":    expense.Source,
		"Amount":    expense.Amount,
		"Currency":  expense.Currency,
		"Notes":     expense.Notes,
		"CreatedAt": expense.CreatedAt,
		"UpdatedAt": expense.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}
	return expense, nil
}

func (d *Database) DeleteExpense(ctx context.Context, userID string, expenseID string) error {
	_, err := d.Firestore.Collection("expenses").Doc(userID).Collection("expenses").Doc(expenseID).Delete(ctx)
	return err
}
