# Keep dot_zshrc.tmpl as a single consolidated file

`dot_zshrc.tmpl` holds ~13 labelled sections in one file. We deliberately keep it
consolidated rather than splitting into a sourced `~/.config/zsh/*.zsh` directory.

The plugin load order is a real invariant: autosuggestions, then
syntax-highlighting (after `compinit`), then history-substring-search (whose
keybindings depend on syntax-highlighting). Keeping every section in one file
makes that ordering visible and local. Splitting would scatter the invariant
across files and make load order implicit and fragile. The file is well under our
size limits, so there is no pressure to split.

This consolidation was itself a deliberate move away from the former
`ansible-desktop` blockinfile sections (see the header comment in the file).
