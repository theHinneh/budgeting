# budgeting
Backend for my budgeting app

## Database backends

This service now supports two database/auth backends:
- Postgres (default)
- Firebase (Auth + Firestore)

Switch the backend using the environment variable `DB_DRIVER`:
- `postgres` (default): uses GORM + Postgres, runs migrations
- `firebase`: initializes Firebase App, Auth, and Firestore; migrations are skipped

### Environment variables

Postgres (already present in `.env`):
- DB_DRIVER=postgres
- DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE, DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME, DB_CONN_TIMEOUT, DB_PG_SCHEMA, DB_PG_SEARCH_PATH, DB_PG_SSLCERT, DB_PG_SSLKEY, DB_PG_SSLROOTCERT

Firebase (add these to `.env` if using Firebase):
- DB_DRIVER=firebase
- FIREBASE_PROJECT_ID=<your_project_id>
- FIREBASE_CREDENTIALS_FILE=<path_to_service_account_json> (optional)

Credentials setup (per official guide):
- You can use Application Default Credentials (ADC). Set GOOGLE_APPLICATION_CREDENTIALS to the path of your service account JSON, or run `gcloud auth application-default login` in dev.
- Alternatively, specify FIREBASE_CREDENTIALS_FILE to explicitly pass the service account file to the SDK.
- Official setup guide: https://firebase.google.com/docs/admin/setup/#go

### Health check

The `/health` endpoint reports the selected backend's connectivity using a lightweight Ping. For Firebase, it touches the Firestore API. For Postgres, it pings the SQL connection.
