#!/usr/bin/env bash
# Offline test of update-vendored-skills.sh against file:// fixtures. The script
# runs unattended under Renovate, so what matters most is how it fails: never
# reporting success over a wiped skill, never touching a locally-authored one.
#
# Requires: curl, jq.
# Run:  bash scripts/test-update-vendored-skills.sh
set -euo pipefail

SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/update-vendored-skills.sh"
REPO="mattpocock/skills"
REF="$(sed -n 's/^MATTPOCOCK_SKILLS_REF="\([0-9a-f]*\)"$/\1/p' "$SCRIPT")"
TODAY="$(date -u +%Y-%m-%d)"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
FX="${WORK}/fixture"
failures=0

build_fixture() {
  local raw="${FX}/raw/${REPO}/${REF}" path
  local paths=(skills/a/alpha/SKILL.md skills/a/beta/SKILL.md skills/a/beta/ref/notes.md skills/b/one/SKILL.md)
  mkdir -p "${FX}/api/repos/${REPO}/git/trees"
  for path in "${paths[@]}"; do
    mkdir -p "${raw}/$(dirname "$path")"
    echo "content of ${path}" >"${raw}/${path}"
  done
  echo "Fixture licence" >"${raw}/LICENSE"
  printf '%s\n' "${paths[@]}" |
    jq -R '{path: ., type: "blob"}' | jq -s '{tree: .}' \
      >"${FX}/api/repos/${REPO}/git/trees/${REF}"
}

# $1 = SKILLS_ROWS, $2 = skills dir. Exit status is the script's.
run_vendor() {
  SKILLS_ROWS="$1" SKILLS_DIR="$2" \
    GITHUB_API="file://${FX}/api" GITHUB_RAW="file://${FX}/raw" \
    bash "$SCRIPT" >"${WORK}/out" 2>&1
}

# $@ as for run_vendor; succeeds only when the script fails.
vendor_fails() {
  # shellcheck disable=SC2310
  ! run_vendor "$@"
}

check() {
  # $1 = description, rest = command that must succeed.
  local what="$1"
  shift
  if "$@"; then
    echo "ok: ${what}"
  else
    echo "FAIL: ${what}"
    sed 's/^/  | /' "${WORK}/out"
    failures=$((failures + 1))
  fi
}

fresh_dir() {
  rm -rf "${WORK}/skills"
  mkdir -p "${WORK}/skills"
  echo "${WORK}/skills"
}

seed_vendored() {
  # $1 = dir, $2 = skill, $3 = pinned ref, $4 = retrieval date
  mkdir -p "$1/$2"
  printf -- "- Pinned commit: \`%s\`\n- Retrieved: %s\n" "$3" "$4" >"$1/$2/UPSTREAM.md"
  echo keep >"$1/$2/marker"
}

ROWS="${REPO}|skills/a|*
${REPO}|skills/b|one"

build_fixture

dir="$(fresh_dir)"
check "wildcard and named rows vendor every file" \
  run_vendor "$ROWS" "$dir"
check "nested file fetched" test -f "${dir}/beta/ref/notes.md"
check "named skill fetched" test -f "${dir}/one/SKILL.md"
check "licence embedded" grep -qF "Fixture licence" "${dir}/alpha/UPSTREAM.md"

dir="$(fresh_dir)"
seed_vendored "$dir" alpha "$REF" 2000-01-01
seed_vendored "$dir" beta 0000000000000000000000000000000000000000 2000-01-01
run_vendor "$ROWS" "$dir"
check "unchanged pin keeps its date" grep -qFx -- "- Retrieved: 2000-01-01" "${dir}/alpha/UPSTREAM.md"
check "changed pin gets today's date" grep -qFx -- "- Retrieved: ${TODAY}" "${dir}/beta/UPSTREAM.md"

dir="$(fresh_dir)"
seed_vendored "$dir" stale "$REF" 2000-01-01
mkdir -p "${dir}/mine" && echo local >"${dir}/mine/SKILL.md"
run_vendor "$ROWS" "$dir"
check "skill no longer produced is pruned" test ! -e "${dir}/stale"
check "locally-authored skill survives pruning" test -f "${dir}/mine/SKILL.md"

dir="$(fresh_dir)"
seed_vendored "$dir" alpha "$REF" 2000-01-01
mkdir -p "${dir}/beta" && echo local >"${dir}/beta/SKILL.md"
check "clash with a local skill fails" vendor_fails "$ROWS" "$dir"
check "clash leaves the local skill as it was" grep -qFx local "${dir}/beta/SKILL.md"
check "clash wipes nothing else" test -f "${dir}/alpha/marker"

dir="$(fresh_dir)"
seed_vendored "$dir" alpha "$REF" 2000-01-01
check "missing upstream tree fails" vendor_fails "${ROWS}
cloudflare/security-audit-skill|skills|security-audit" "$dir"
check "failed fetch wipes nothing" test -f "${dir}/alpha/marker"

dir="$(fresh_dir)"
check "same name from two rows fails" vendor_fails "${ROWS}
${REPO}|skills/a|alpha" "$dir"

if ((failures > 0)); then
  echo "${failures} check(s) failed" >&2
  exit 1
fi
echo "all checks passed"
