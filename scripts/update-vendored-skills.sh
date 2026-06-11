#!/usr/bin/env bash
# Re-vendor the community Claude skills from upstream at a pinned commit.
#
# The pinned ref below is bumped by Renovate (git-refs digest datasource).
# A Renovate postUpgradeTask then re-runs this script so the same PR carries
# the refreshed skill content, not just a new SHA string.
#
# The file set for each skill is derived from the upstream tree at the pinned
# ref, not hardcoded here: only the skill names are curated. A file added or
# removed upstream is mirrored on the next run without editing this script.
#
# Requires: curl, jq.
# Run manually:  bash scripts/update-vendored-skills.sh
set -euo pipefail

# renovate: datasource=git-refs depName=mattpocock/skills packageName=https://github.com/mattpocock/skills branch=main
UPSTREAM_REF="694fa30311e02c2639942308513555e61ee84a6f"

UPSTREAM_REPO="mattpocock/skills"
UPSTREAM_SUBPATH="skills/engineering"
RAW_BASE="https://raw.githubusercontent.com/${UPSTREAM_REPO}/${UPSTREAM_REF}/${UPSTREAM_SUBPATH}"
TREE_URL="https://api.github.com/repos/${UPSTREAM_REPO}/git/trees/${UPSTREAM_REF}?recursive=1"
SKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/dot_claude/skills"
RETRIEVED="$(date -u +%Y-%m-%d)"

# The curated set. Their file lists are read from upstream, not stated here.
SKILLS=(grill-with-docs improve-codebase-architecture)

# Fetch the upstream tree once and reuse it to list each skill's files.
TREE_JSON="$(curl -fsSL "$TREE_URL")"

write_upstream_md() {
  # $1 = skill name, $2 = destination dir
  cat > "$2/UPSTREAM.md" <<EOF
# Upstream provenance

Vendored from [${UPSTREAM_REPO}](https://github.com/${UPSTREAM_REPO}) —
\`${UPSTREAM_SUBPATH}/$1\`.

- Pinned commit: \`${UPSTREAM_REF}\`
- Retrieved: ${RETRIEVED}
- Refreshed by: \`scripts/update-vendored-skills.sh\` (bumped by Renovate)

## License

Upstream is distributed under the MIT License:

\`\`\`
MIT License

Copyright (c) 2026 Matt Pocock

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
\`\`\`
EOF
}

list_skill_files() {
  # $1 = skill name -> echoes each file path relative to the skill's dir.
  jq -r --arg p "${UPSTREAM_SUBPATH}/$1/" \
    '.tree[] | select(.type == "blob") | select(.path | startswith($p)) | .path[($p | length):]' \
    <<<"$TREE_JSON"
}

vendor_skill() {
  # $1 = skill name
  local skill="$1" dest f
  dest="${SKILLS_DIR}/${skill}"
  # Wipe and refetch so a file removed upstream does not linger. UPSTREAM.md is
  # regenerated below, so nothing hand-written is lost in the wipe.
  rm -rf "$dest"
  mkdir -p "$dest"
  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    echo "fetching ${skill}/${f}"
    mkdir -p "${dest}/$(dirname "$f")"
    curl -fsSL "${RAW_BASE}/${skill}/${f}" -o "${dest}/${f}"
  done < <(list_skill_files "$skill")
  write_upstream_md "$skill" "$dest"
}

for skill in "${SKILLS[@]}"; do
  vendor_skill "$skill"
done

echo "done: vendored skills refreshed at ${UPSTREAM_REF}"
