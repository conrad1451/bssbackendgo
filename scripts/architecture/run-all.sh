#!/usr/bin/env bash

# CHQ: Gemini AI created file

set -euo pipefail

SCHEMA_PATH="tmp/schema.sql"
export SCHEMA_PATH

echo "🚦 Running architecture enforcement checks..."

sh scripts/architecture/00_dump_schema.sh

for script in scripts/architecture/*.sh; do
  case "$(basename "$script")" in
    00_dump_schema.sh|run-all.sh) continue ;;
    *) sh "$script" ;;
  esac
done

echo "🎉 All architecture checks passed"
