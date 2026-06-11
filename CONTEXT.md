# dotfiles

chezmoi-managed macOS configuration. This file records the project-specific
vocabulary so reviews and changes use one consistent set of names.

## Language

**Machine profile**:
The set of per-machine values prompted once at `chezmoi init` and stored only in
the local `~/.config/chezmoi/chezmoi.toml` — never committed. Defined in exactly
one place: the `promptStringOnce` calls in `.chezmoi.toml.tmpl`. Both consumers
(interactive init and the CI render-check) derive from this single definition.
_Avoid_: prompts, variables, secrets, config data.

## Example dialogue

> **Dev:** CI broke after I added a Docker socket prompt — do I need to update the
> Woodpecker step?
> **Maintainer:** No. The machine profile is the single definition; CI runs
> `chezmoi init --promptDefaults`, so any prompt with a default renders with that
> default automatically. You only edit `.chezmoi.toml.tmpl`.
