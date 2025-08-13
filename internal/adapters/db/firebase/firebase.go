package firebase

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	fb "firebase.google.com/go/v4"
	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/theHinneh/budgeting/pkg/config"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type Database struct {
	app       *fb.App
	Firestore *firestore.Client
	Auth      *fbAuth.Client
}

// NewDatabase initializes Firebase App and clients using env/config.
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

func (d *Database) GetDB() *gorm.DB { return nil }

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
