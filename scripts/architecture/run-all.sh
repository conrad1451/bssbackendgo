# CHQ: Gemini AI created file

#!/usr/bin/env sh
set -e

echo "🚦 Running architecture enforcement checks..."

for script in scripts/architecture/*.sh; do
  if [ "$(basename "$script")" != "run-all.sh" ]; then
    sh "$script"
  fi
done

echo "🎉 All architecture checks passed"
