#!/usr/bin/env bash
set -e

# Fail if token.ID is used outside middleware
if rg "token\.ID" . \
  | rg -v "sessionValidationMiddleware" \
  | rg -v "docs/" \
  | rg -v "vendor/"; then
  echo "❌ token.ID used outside sessionValidationMiddleware"
  exit 1
fi

echo "✅ token.ID usage OK"
