# ADR-0007: Vendor whole upstream categories, not hand-picked skills

Status: Accepted. Supersedes the "only the skill names are curated" part of ADR-0006.

## Context

ADR-0006 curated mattpocock/skills by name: 6 of the 39 skills at `4ae2317`. The
chain upstream is built around (`grill-with-docs`, `to-spec`, `to-tickets`,
`implement`, `code-review`, `retro`) calls skills across categories: `retro` loads
`writing-for-agents` first. A partial pick breaks those calls silently, and each
new upstream skill needed a script edit before it could even be evaluated.

## Decision

A `SKILLS` row may name `*` as its skill, meaning every directory with a
`SKILL.md` directly under that subpath at the pinned SHA.

- mattpocock/skills: `engineering/`, `misc/`, `productivity/`. Not
  `in-progress/` or `deprecated/`, which upstream marks as not ready.
- DietrichGebert/ponytail: all of `skills/`.
- Other repositories stay curated by name.

Everything else in ADR-0006 stands: pinned SHAs, Renovate bumps, `UPSTREAM.md`
with the verbatim licence.

## Consequences

- A new upstream skill arrives in a Renovate PR without a script edit. Review
  moves from "pick before" to "read the diff after".
- A vendored directory no upstream row produces is pruned, so a skill removed
  upstream does not linger. `UPSTREAM.md` is the marker: a directory without it
  is locally authored, never pruned, and never overwritten. A name clash with
  one aborts the run before anything is wiped.
- Vendored `code-review` replaces the bundled `/code-review`. Claude Code gives a
  personal skill precedence over a bundled one; the bundled `/review` alias
  still runs the built-in.
- `.pre-commit-config.yaml` excludes vendored skills by listing the
  locally-authored ones in a negative lookahead, since the vendored set is no
  longer known in advance. A new local skill must be added there.
- The 15 or so model-invocable skills add their descriptions to every session's
  context; the rest are `disable-model-invocation` and cost nothing until called.
