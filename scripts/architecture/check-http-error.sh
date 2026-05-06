#!/usr/bin/env bash
set -e

if rg "http\.Error\(" . \
  --glob '!scripts/' \
  --glob '!*_test.go' \
  --glob '!vendor/'; then
  echo "❌ http.Error used — use writeJSONError"
  exit 1
fi

echo "✅ error handling OK"
