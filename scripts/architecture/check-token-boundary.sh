#!/usr/bin/env bash
set -e

if rg "token\.ID" . \
  | rg -v "sessionValidationMiddleware" \
  | rg -v "docs/" \
  | rg -v "vendor/"; then
  echo "❌ token.ID used outside sessionValidationMiddleware"
  exit 1
fi

echo "✅ token.ID usage OK"
