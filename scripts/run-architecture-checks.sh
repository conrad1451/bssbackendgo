#!/usr/bin/env bash
set -e

echo "🔍 Running architecture checks..."

for check in scripts/arch/*.sh; do
  echo "▶ $check"
  bash "$check"
done

echo "✅ All architecture checks passed"
