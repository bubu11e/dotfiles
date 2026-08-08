# ADR-0005: `TODO.md` is machine-local and gitignored

Status: Accepted

## Context

The global working agreement for these repositories says to update and commit a
TODO file alongside the code it describes. That rule exists so a shared task list
does not drift from the work.

This repository's `TODO.md` is not that kind of file. It holds half-formed
roadmap notes, findings from reading other people's dotfiles, and a "decided
against" section that exists to stop old arguments being reopened. Committing it
would put every unfinished thought into a public history and turn each exploratory
edit into a commit that needs a message.

## Decision

`TODO.md` lives at the repo root, is listed in `.gitignore`, and is never staged.
This is a deliberate, recorded exception to the global rule.

Do not propose tracking it again.

## Consequences

- Edit it freely; never `git add` it.
- It is machine-local. It does not survive a fresh clone, and that is accepted —
  anything that must survive belongs in an ADR, in `CONTEXT.md`, or in a comment
  next to the code it constrains.
- Being at the root, it is also outside `home/` (ADR-0001), so it can never be
  deployed into `$HOME` either.
- A decision that gets *made* in `TODO.md` graduates: it moves into an ADR here,
  and the TODO entry goes away. The "decided against" section is for calls not
  worth a full record.
