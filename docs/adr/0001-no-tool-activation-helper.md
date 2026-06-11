# No shared helper for guarded tool activation

The `command -v X && activate` guard in `dot_zshrc.tmpl` is repeated across fzf,
mise, kubectl, and starship. We deliberately keep these guards inline rather than
extracting a `_activate <cmd> <init-string>` helper.

A helper would be shallow: it covers only the single-line activations (homebrew,
golang, and the load-order-coupled plugins block cannot use it) and forces an
`eval "$2"` of a command string, trading readable inline guards for quoting
hazards. The duplication is three or four lines of an idiomatic, well-understood
shell pattern; concentrating it does not earn its keep.

The real defect the review found was not duplication but two snippets (starship,
gpg) that lacked the guard the README documents — those were fixed in place.

## Amendment: the completion case is different

A later change *did* extract a `_tool_completion <cmd>` helper for tools that
expose `<cmd> completion zsh` (docker, podman). This does not reopen the
decision above. The activation helper was rejected because it forced an
`eval "$2"` of a command string (quoting hazards) and covered only a couple of
single-line cases. The completion helper has neither problem: it is a plain
`source <($1 completion zsh)` with no command-string eval, and the call sites
share the exact same guard-and-source shape — a real seam, not a hypothetical
one. kubectl stays inline because it needs an alias + `compdef` alongside its
completion, so a single guard covers both without probing the binary twice. The
activation guards (homebrew, mise, fzf, starship) remain inline.
