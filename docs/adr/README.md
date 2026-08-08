# Architecture decision records

Why this repository is shaped the way it is. Each record states the forces at the
time and the call that was made, so a future change is an argument against a known
position rather than a rediscovery.

These are decisions, not documentation. How to *use* the setup is in `README.md`;
how to *validate* a change to it is in `CLAUDE.md`; what is still open is in
`TODO.md` (machine-local, see ADR-0005).

Format: context, decision, consequences. Short. A record that needs more than a
page is describing several decisions.

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-chezmoiroot-home.md) | Source directory is `home/`, set by `.chezmoiroot` | Accepted |
| [0002](0002-shell-config-layering.md) | Layer shell config by when the shell reads it | Accepted |
| [0003](0003-branch-on-ostype.md) | Branch on `$OSTYPE`, never on `uname` | Accepted |
| [0004](0004-completions-from-fpath.md) | Take completions from `$fpath`, not from a startup fork | Accepted |
| [0005](0005-todo-is-machine-local.md) | `TODO.md` is machine-local and gitignored | Accepted |
| [0006](0006-vendor-skills-by-pinned-sha.md) | Vendor upstream skills at pinned SHAs, by script | Accepted |

## Superseding

Amend a record in place only for typos. A reversal gets a new record that says
which one it supersedes, and the old record's status becomes `Superseded by ADR-N`.
The point is the trail, not the tidiness.
