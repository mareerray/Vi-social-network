#!/usr/bin/env bash
set -euo pipefail

DB_FILE=${DB_PATH:-/app/backend/socialnetwork.db}
MIGRATIONS_DIR=/migrations

echo "Entrypoint: ensuring DB file exists at $DB_FILE"
mkdir -p "$(dirname "$DB_FILE")"
if [ ! -f "$DB_FILE" ]; then
  echo "Creating new SQLite DB file"
  sqlite3 "$DB_FILE" ".databases"
fi

echo "Applying migrations from $MIGRATIONS_DIR"
if [ -d "$MIGRATIONS_DIR" ]; then
  for f in $(ls -1 "$MIGRATIONS_DIR"/*up.sql 2>/dev/null | sort); do
    echo "Applying migration: $f"
    # read migration SQL into a single line for lightweight parsing
    sql=$(tr '\n' ' ' < "$f" | tr -s ' ')

    # If this migration is an ALTER TABLE ... ADD COLUMN, skip if column already exists
    if echo "$sql" | grep -Eiq 'ALTER TABLE[[:space:]]+[a-zA-Z0-9_]+[[:space:]]+ADD COLUMN[[:space:]]+[a-zA-Z0-9_]+'; then
      table=$(echo "$sql" | sed -nE 's/.*ALTER TABLE[[:space:]]+([a-zA-Z0-9_]+).*/\1/p' | tr -d '\r')
      column=$(echo "$sql" | sed -nE 's/.*ADD COLUMN[[:space:]]+([a-zA-Z0-9_]+).*/\1/p' | tr -d ',;\r')
      if [ -n "$table" ] && [ -n "$column" ]; then
        # Check if the column exists in the table (PRAGMA table_info returns rows where column name is $2)
        col_exists=$(sqlite3 "$DB_FILE" "PRAGMA table_info($table);" | awk -F'|' -v c="$column" '{ if ($2 == c) print "yes" }' || true)
        if [ "$col_exists" = "yes" ]; then
          echo "Skipping migration $f: column '$column' already exists on table '$table'"
          continue
        fi
      fi
    fi

    # If this migration creates a table, skip if table already exists
    if echo "$sql" | grep -Eiq 'CREATE TABLE[[:space:]]+[a-zA-Z0-9_]+'; then
      table=$(echo "$sql" | sed -nE 's/.*CREATE TABLE[[:space:]]+([a-zA-Z0-9_]+).*/\1/p' | tr -d '\r')
      if [ -n "$table" ]; then
        tbl_exists=$(sqlite3 "$DB_FILE" "SELECT name FROM sqlite_master WHERE type='table' AND name='$table';")
        if [ -n "$tbl_exists" ]; then
          echo "Skipping migration $f: table '$table' already exists"
          continue
        fi
      fi
    fi

    sqlite3 "$DB_FILE" < "$f" || { echo "Migration failed: $f"; exit 1; }
  done
else
  echo "No migrations directory found at $MIGRATIONS_DIR"
fi

echo "Starting application"
exec "$@"
