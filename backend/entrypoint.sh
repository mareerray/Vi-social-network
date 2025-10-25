#!/bin/sh
set -e

DB_FILE=${DB_PATH:-/app/socialnetwork.db}
MIGRATIONS_DIR=/app/db/migrations/sqlite

echo "=== Starting Social Network Backend ==="
echo "Database: $DB_FILE"

# Ensure DB directory exists
mkdir -p "$(dirname "$DB_FILE")"

# Create DB if it doesn't exist
if [ ! -f "$DB_FILE" ]; then
  echo "Creating new SQLite database..."
  sqlite3 "$DB_FILE" "SELECT 1;" > /dev/null 2>&1 || true
fi

# Apply migrations
echo "Checking for migrations in $MIGRATIONS_DIR"
if [ -d "$MIGRATIONS_DIR" ]; then
  echo "Applying migrations..."
  for migration in $(ls -1 "$MIGRATIONS_DIR"/*up.sql 2>/dev/null | sort); do
    migration_name=$(basename "$migration")
    echo "  → $migration_name"
    
    # Simple approach: try to apply, ignore errors if already applied
    sqlite3 "$DB_FILE" < "$migration" 2>/dev/null || {
      echo "    (already applied or error - continuing)"
    }
  done
  echo "Migrations complete"
else
  echo "Warning: No migrations directory found at $MIGRATIONS_DIR"
fi

echo "Starting application..."
exec "$@"
