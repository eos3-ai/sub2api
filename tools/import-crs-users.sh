#!/usr/bin/env bash
set -euo pipefail

# Wrapper for `tools/import-crs-users.psql`.
#
# Usage:
#   tools/import-crs-users.sh /path/to/sub2api-users.csv
#
# Env:
#   DATABASE_URL   Optional. If set, used as psql connection string.
#   CONCURRENCY    Optional. Default 5.
#   ADMIN_EMAIL    Optional. If set, excluded from import.

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 /path/to/sub2api-users.csv" >&2
  exit 2
fi

FILE="$1"
if [[ ! -f "$FILE" ]]; then
  echo "File not found: $FILE" >&2
  exit 2
fi
if [[ "$FILE" != /* ]]; then
  FILE="$(pwd)/$FILE"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PSQL_SCRIPT="${SCRIPT_DIR}/import-crs-users.psql"
if [[ ! -f "$PSQL_SCRIPT" ]]; then
  echo "psql script not found: $PSQL_SCRIPT" >&2
  exit 2
fi

CONCURRENCY="${CONCURRENCY:-5}"
ADMIN_EMAIL="${ADMIN_EMAIL:-}"

PSQL_VARS=(-v "concurrency=${CONCURRENCY}")
if [[ -n "$ADMIN_EMAIL" ]]; then
  PSQL_VARS+=(-v "admin_email=${ADMIN_EMAIL}")
fi

DRY_RUN="${DRY_RUN:-0}"

# Escape the file path for a SQL single-quoted literal.
# (This is only used for the local client-side \copy filename, not sent to the server.)
FILE_ESCAPED="$(printf '%s' "$FILE" | sed "s/'/''/g")"

END_TX="COMMIT;"
if [[ "$DRY_RUN" == "1" ]]; then
  END_TX="ROLLBACK;"
fi

PSQL_ARGS=(-X)
if [[ -n "${DATABASE_URL:-}" ]]; then
  PSQL_ARGS+=("${DATABASE_URL}")
fi

psql "${PSQL_ARGS[@]}" "${PSQL_VARS[@]}" <<SQL
\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE crs_users_import (
  email TEXT NOT NULL,
  usernames TEXT,
  username TEXT,
  password_hash TEXT NOT NULL,
  available_balance NUMERIC NOT NULL
);

\copy crs_users_import(email, usernames, username, password_hash, available_balance) FROM '${FILE_ESCAPED}' WITH (FORMAT csv, HEADER true)

\i ${PSQL_SCRIPT}

$END_TX
SQL
