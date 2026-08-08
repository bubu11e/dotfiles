# ADR-0003: Branch on `$OSTYPE`, never on `uname`

Status: Accepted

## Context

The shared shell files need to tell macOS from Linux in a few places: `ls --color`
versus `ls -G`, whether to set `LANG`, which Homebrew prefix to look for.

`uname -s` is the reflex. It is also a fork, and these files are sourced by every
shell — including the non-interactive ones a script spawns in a loop. A fork that
costs a millisecond is free once and expensive when it is on the path of every
shell start on the machine.

Both bash and zsh set `$OSTYPE` themselves, before any config runs.

## Decision

Branch on `${OSTYPE:-}` with a `case`. Never call `uname` in shell config.

```sh
case "${OSTYPE:-}" in
    linux*)  alias ls='ls --color=auto' ;;
    darwin*) alias ls='ls -G' ;;
esac
```

## Consequences

- Zero forks for a decision that has to be made several times per shell start.
- Patterns need a trailing `*`: the value is `darwin24`, `linux-gnu`, not a bare
  OS name.
- `${OSTYPE:-}` is written with the default expansion because the shared files run
  under `set -u` in some contexts, and because a POSIX `sh` that is neither bash
  nor zsh would not set it at all — such a shell falls through every branch, which
  is the correct degradation.
- The same reasoning bans probing a tool to learn about it (running `ls --color`
  to see whether it errors costs two forks). Ask the shell, not the system.
