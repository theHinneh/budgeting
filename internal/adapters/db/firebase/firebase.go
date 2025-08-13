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

// Database implements ports.DatabasePort for Firebase-backed services.
// It provides Ping and Close. AutoMigrate is a no-op and GetDB returns nil.
// This allows the rest of the app (like Health) to work without changing
// the existing Postgres plumbing.
//
// Note: For actual data access, create repositories that use this adapter's
// Firestore or Auth clients directly.

type Database struct {
	app       *fb.App
	Firestore *firestore.Client
	Auth      *fbAuth.Client
}

// NewDatabase initializes Firebase App and clients using env/config.
// It supports credentials via FIREBASE_CREDENTIALS_FILE or Application Default Credentials.
func NewDatabase(ctx context.Context, cfg *config.Configuration) (*Database, error) {
	projectID := cfg.V.GetString("FIREBASE_PROJECT_ID")
	credsFile := cfg.V.GetString("FIREBASE_CREDENTIALS_FILE")
	srvsAccID := cfg.V.GetString("FIREBASE_SA_ID")

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

// AutoMigrate is a no-op for Firestore.
func (d *Database) AutoMigrate(models ...interface{}) error { return nil }

// GetDB returns nil as GORM is not used with Firebase.
func (d *Database) GetDB() *gorm.DB { return nil }

// Close closes Firestore client.
func (d *Database) Close() error {
	if d.Firestore != nil {
		return d.Firestore.Close()
	}
	return nil
}

// Ping performs a lightweight Firestore operation to verify connectivity.
func (d *Database) Ping(ctx context.Context) error {
	if d.Firestore == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	// Attempt to read zero docs from a non-invasive query (noop but touches the API)
	iter := d.Firestore.Collections(ctx)
	// Just try to iterate one step to ensure API responds
	_, err := iter.Next()
	if err != nil && err != iterator.Done {
		// Any error other than iterator.Done indicates a connectivity/config problem.
		return err
	}

	return nil
}
