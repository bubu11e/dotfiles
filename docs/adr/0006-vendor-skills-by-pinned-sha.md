# ADR-0006: Vendor upstream skills at pinned SHAs, by script

Status: Accepted

## Context

Several Claude skills used here are written by other people. There are two ways to
consume them: install them declaratively as plugins from a marketplace, letting the
tool fetch whatever is current, or copy the files into this repository.

Declarative is less code. It also means the content that shapes an agent's
behaviour arrives from a third party without passing through review, without a
diff, and without the gitleaks scan every other file here gets. The upgrade is
invisible: nothing in this repository changes when upstream changes.

## Decision

Vendor the files into `home/dot_claude/skills/`, fetched by
`scripts/update-vendored-skills.sh` from a SHA pinned per upstream repository.

- Only the skill *names* are curated. Each skill's file list is read from the
  upstream tree at the pinned SHA, so a file added or removed upstream is mirrored
  without editing the script.
- Renovate bumps each pinned SHA, and a `postUpgradeTask` re-runs the script so the
  same pull request carries the refreshed content rather than just a new SHA
  string.
- Each skill gets an `UPSTREAM.md` recording its origin, pinned commit, retrieval
  date, and a verbatim copy of the upstream `LICENSE` at that commit.

## Consequences

- Every upstream change arrives as a reviewable diff and is scanned by gitleaks
  like any other file.
- The vendored tree is third-party content: `.yamllint` ignores it, because we do
  not control its style and the next refetch would revert any local fix.
- Local edits to a vendored skill are pointless — the script wipes and refetches
  the directory. A change that must persist belongs upstream or in a
  locally-authored skill.
- The script runs unattended under Renovate, which sets the bar for its failure
  behaviour: it must never leave a skill partially or wholly deleted and still
  report success. That is a real bug that was found and fixed, and it is why the
  fetch helpers propagate failure explicitly instead of relying on `set -e`
  through nested command substitutions.
- Pinning by SHA means the "cloudflare" `_REF` variable is shape-identical to a
  Cloudflare API key. `.gitleaks.toml` carries a narrowly scoped allowlist for
  exactly that rule, that file, and values matching a 40-hex ref.
