#!/usr/bin/env bash
# Fails when a Go file is over 500 lines. Files that were already over the
# limit are listed in file-length-baseline.txt with their length at the time;
# they may shrink but not grow.
set -euo pipefail
cd "$(dirname "$0")/.."

limit=500
baseline=scripts/file-length-baseline.txt
status=0

while IFS= read -r f; do
  n=$(wc -l < "$f" | tr -d ' ')
  allowed=$(awk -v f="$f" '$2 == f { print $1 }' "$baseline")
  if [ -n "$allowed" ] && [ "$n" -gt "$allowed" ]; then
    echo "$f: $n lines, up from its baseline of $allowed. It is already over the $limit-line limit, so put new code in a separate file split by responsibility (for example, rendering in view.go) instead of growing this one."
    status=1
  elif [ -z "$allowed" ] && [ "$n" -gt "$limit" ]; then
    echo "$f: $n lines, over the $limit-line limit. Split it into smaller files by responsibility (for example, rendering in view.go, data tables in their own file)."
    status=1
  fi
done < <(git ls-files --cached --others --exclude-standard '*.go')

exit $status
