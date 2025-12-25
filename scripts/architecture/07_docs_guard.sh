# CHQ: Gemini AI created file

#!/usr/bin/env sh
set -e

echo "🔍 Checking architecture documentation..."

if [ ! -f docs/architecture.md ]; then
  echo "❌ docs/architecture.md missing"
  exit 1
fi

# Ensure key invariants are documented
REQUIRED_STRINGS="
Descope IDs are never used as database primary keys
sessionValidationMiddleware
authorization checks must use players.id
"

for s in $REQUIRED_STRINGS; do
  if ! grep -F "$s" docs/architecture.md >/dev/null; then
    echo "❌ Missing invariant in architecture.md: $s"
    exit 1
  fi
done

echo "✅ Documentation checks passed"

