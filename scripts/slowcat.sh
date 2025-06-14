#!/usr/bin/env bash
# slowcat.sh – print a file to STDOUT one line per second
# Usage:  ./scripts/slowcat.sh <file>
set -euo pipefail
if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <file>" >&2
  exit 1
fi
file="$1"
if [[ ! -f "$file" ]]; then
  echo "File not found: $file" >&2
  exit 1
fi
while IFS= read -r line || [[ -n "$line" ]]; do
  echo "$line"
  sleep 1
done < "$file"
