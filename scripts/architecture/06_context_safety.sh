# CHQ: Gemini AI created file

#!/usr/bin/env sh
set -e

echo "🔍 Checking context safety..."

# Disallow string context keys
if grep -R "context.WithValue(.*, \".*\"" .; then
  echo "❌ String-based context key detected"
  exit 1
fi

# Ensure custom contextKey type exists
if ! grep -R "type contextKey" . >/dev/null; then
  echo "❌ Custom contextKey type not defined"
  exit 1
fi

echo "✅ Context safety checks passed"
