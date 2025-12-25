#!/usr/bin/env bash
set -e

if rg "http\.Error\(" . \
  | rg -v "_test.go" \
  | rg -v "vendor/"; then
  echo "❌ http.Error used — use writeJSONError"
  exit 1
fi

echo "✅ error handling OK"
