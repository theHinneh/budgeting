package firebase

import (
	"context"
	"fmt"

	fb "firebase.google.com/go/v4"
	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/theHinneh/budgeting/internal/infrastructure/config"
	"google.golang.org/api/option"

	"cloud.google.com/go/firestore"
)

type Database struct {
	app             *fb.App
	FirestoreClient *firestore.Client
	AuthClient      *fbAuth.Client

	UserRepository         *UserRepository
	UserAuthenticator      *FirebaseAuth
	IncomeRepository       *IncomeRepository
	ExpenseRepository      *ExpenseRepository
	IncomeSourceRepository *IncomeRepository
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
	srvsAccID := getStr("FIREBASE_SA_ID", "database.FIREBASE_SA_ID", "database.firebase_s-id")

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

	return &Database{
		app:                    app,
		FirestoreClient:        fsClient,
		AuthClient:             authClient,
		UserRepository:         &UserRepository{Firestore: fsClient},
		UserAuthenticator:      &FirebaseAuth{Auth: authClient},
		ExpenseRepository:      &ExpenseRepository{Firestore: fsClient},
		IncomeRepository:       &IncomeRepository{Firestore: fsClient},
		IncomeSourceRepository: &IncomeRepository{Firestore: fsClient},
	}, nil
}

func (d *Database) AutoMigrate(models ...interface{}) error { return nil }

func (d *Database) Close() error {
	if d.FirestoreClient != nil {
		return d.FirestoreClient.Close()
	}
	return nil
}
