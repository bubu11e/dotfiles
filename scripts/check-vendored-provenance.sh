#!/bin/sh
# Every vendored skill must ship its provenance and upstream licence (ADR-0006).
# 31 skills once reached main with an empty UPSTREAM.md: `git checkout` on an
# intent-to-add path (`git add -N`) restores it to empty, and nothing failed.
set -eu

status=0
for f in home/dot_claude/skills/*/UPSTREAM.md; do
  [ -e "$f" ] || continue
  # shellcheck disable=SC2016 # the backticks are literal Markdown, not expansions
  if ! grep -q '^- Pinned commit: `[0-9a-f]\{40\}`$' "$f" || ! grep -q '^## License$' "$f"; then
    echo "missing provenance or licence: $f" >&2
    status=1
  fi
done
exit "$status"
