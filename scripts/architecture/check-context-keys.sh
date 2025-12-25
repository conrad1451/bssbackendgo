#!/usr/bin/env bash
set -e

if rg 'context\.WithValue\(.*, "[^"]+"' .; then
  echo "❌ string context key detected — use typed contextKey"
  exit 1
fi

echo "✅ context keys OK"
