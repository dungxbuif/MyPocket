# Backup and Restore Drill

MyPocket requires a successful PostgreSQL and private-object restore drill before production deployment. Never target the production database during a restore test.

## Create a backup

Load production secrets into the shell without printing them, choose a new explicit directory, then run:

```bash
export MYPOCKET_BACKUP_DIR=/secure/path/mypocket-YYYYMMDDTHHMMSSZ
./scripts/backup-production.sh
```

The script accepts the production names `MYPOCKET_DATABASE_URL`, `MYPOCKET_S3_ENDPOINT`, `MYPOCKET_S3_BUCKET`, `MYPOCKET_S3_REGION`, `MYPOCKET_S3_ACCESS_KEY_ID`, and `MYPOCKET_S3_SECRET_ACCESS_KEY`. The shorter backend names are supported as fallbacks. It requires the `pg_dump` major version to match the source PostgreSQL server, then creates a custom-format dump, deterministic row counts, an S3 inventory, local object copies, tool-version evidence, and SHA-256 manifests under a mode-`0700` directory with mode-`0600` evidence files. It refuses to overwrite an existing backup.

Copy the directory to encrypted storage after creation. The checksums detect corruption but do not encrypt the backup by themselves.

## Run an isolated restore

Create a disposable database whose name starts with `mypocket_restore_`. Choose a dedicated object prefix starting with `restore-drill/`; the script never restores to the bucket root.

```bash
export MYPOCKET_RESTORE_DATABASE_URL='postgres://…/mypocket_restore_20260911?sslmode=require'
export MYPOCKET_RESTORE_S3_PREFIX='restore-drill/20260911T010203Z'
./scripts/restore-drill.sh
```

The drill rejects the source database URL, requires the backup, `pg_restore`, and target PostgreSQL major versions to match, restores with owner/privilege changes disabled, compares every public-table row count, uploads objects only beneath the isolated prefix, downloads them again, and compares their SHA-256 manifest. A successful run writes `restore-evidence.txt` without credentials.

After reviewing the evidence, remove the disposable database and prefix using the infrastructure owner's normal retention procedure. Deletion is intentionally not automated by the drill.
