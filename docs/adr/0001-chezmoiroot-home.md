# ADR-0001: Source directory is `home/`, set by `.chezmoiroot`

Status: Accepted

## Context

A dotfiles repository holds two kinds of file that must never be confused: things
that belong in `$HOME`, and things that describe the repository itself — README,
CI config, lint config, ADRs, helper scripts.

chezmoi's default is to treat the repository root as the source directory. Under
that default every repo-meta file is a candidate for deployment, and the only thing
standing between `README.md` and `~/README.md` is a correctly maintained
`.chezmoiignore`. That file is a denylist, so the failure mode is silent and
arrives with the *next* file somebody adds, not with the change that broke it.

## Decision

`.chezmoiroot` contains `home`. chezmoi's source directory is `home/`, and
everything outside it is invisible to `chezmoi apply`.

There is deliberately no `.chezmoiignore`.

## Consequences

- Deploying a new file means putting it under `home/`. Adding a repo-meta file
  means putting it at the root and doing nothing else. The safe case is the
  default, and the unsafe case requires a deliberate act.
- The rule is structural rather than enumerated, so it cannot rot. No file has to
  be added to a list to stay out of `$HOME`.
- `home/.chezmoi.toml.tmpl` has to live inside `home/`, not at the root, because
  chezmoi looks for it in the source directory.
- Commands run from the repo root need `--source`, which is why CI passes
  `--source="$CI_WORKSPACE"` and the helper scripts compute the repo root
  themselves.
