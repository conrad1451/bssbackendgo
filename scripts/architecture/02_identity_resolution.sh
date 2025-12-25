# CHQ: Gemini AI created file

#!/usr/bin/env sh
set -e

echo "🔍 Checking identity resolution boundaries..."

# Handlers must not parse JWTs
if grep -R "ValidateSession\|ValidateRoles" ./handlers; then
  echo "❌ Handlers must not call Descope directly"
  exit 1
fi

# Handlers must not parse Authorization headers
if grep -R "Authorization" ./handlers; then
  echo "❌ Handlers must not read Authorization header"
  exit 1
fi

# Ensure middleware exists
if ! grep -R "sessionValidationMiddleware" . >/dev/null; then
  echo "❌ sessionValidationMiddleware not found"
  exit 1
fi

echo "✅ Identity resolution checks passed"
