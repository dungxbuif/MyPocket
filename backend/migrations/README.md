# Stage v1 migrations

Stage v1 starts with one baseline migration pair and uses additive forward migrations for post-baseline changes:

- `000001_stage_v1.up.sql` creates the complete schema and canonical seed data.
- `000001_stage_v1.down.sql` is intentionally non-destructive; user data must not be dropped by an automated rollback.
- `000002_ai_entry_error_code.up.sql` adds safe provider diagnostics to one-shot AI entry records.
- `000003_feedback_screenshot.up.sql` adds private screenshot metadata for owner-scoped feedback.

The previous incremental migration files were squashed for the stage reset. Existing databases whose migration table is at version 19 must be recreated or explicitly reset to version 1 in a disposable environment before running this baseline. Do not force or drop a production database.

Local reset example (only for a disposable database):

```sh
dropdb --if-exists mypocket_stage_v1
createdb mypocket_stage_v1
DATABASE_URL='postgres://dev:password@127.0.0.1:5432/mypocket_stage_v1?sslmode=disable' MIGRATIONS_PATH=backend/migrations go run ./backend/cmd/migrate up
```

The application and tests continue to use `MIGRATIONS_PATH` when a non-default migration directory is needed.
