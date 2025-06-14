#!/usr/bin/env bash
# simulate_cache.sh – stream coloured cache restore/save log lines to STDOUT
# A simple way to preview Buildkite cache plugin output with terminal-to-html.
#
# By default it reads data/input.json and emits two lines per entry, pausing
# one second between lines to simulate a live build.
#
# Usage:
#   ./scripts/simulate_cache.sh              # uses data/input.json
#   ./scripts/simulate_cache.sh myfile.json  # use custom fixture
#
set -euo pipefail
json_file="${1:-data/input.json}"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required (brew install jq or apt-get install jq)" >&2
  exit 1
fi

if [[ ! -f "$json_file" ]]; then
  echo "JSON file not found: $json_file" >&2
  exit 1
fi

jq -c '.[]' "$json_file" | while read -r entry; do
  action=$(jq -r '.action' <<<"$entry")    # restore | save
  cache=$(jq -r '.cache'  <<<"$entry")
  duration=$(jq -r '.duration' <<<"$entry")
  hit=$(jq -r '.cache_hit // false' <<<"$entry")

  icon="♻️"; colour=34               # blue for restore
  if [[ "$action" == "save" ]]; then icon="💾"; colour=35; fi

  # Line 1 – start of step
  printf '\e[1m\e[%sm%s %s\e[0m  cache=%s\n' "$colour" "$icon" "$action" "$cache"
  sleep 1

  # Line 2 – result
  if [[ "$action" == "restore" ]]; then
    if [[ "$hit" == true ]]; then
      printf '\e[32m🎯 Hit\e[0m  duration=%s\n' "$duration"
    else
      printf '\e[31m💨 Miss\e[0m duration=%s\n' "$duration"
    fi
  else # save
    printf '\e[36mSaved\e[0m   duration=%s\n' "$duration"
  fi
  sleep 1

done
