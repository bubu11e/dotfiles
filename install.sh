#!/bin/sh
# Bootstrap this checkout onto the current machine: install chezmoi if it is
# absent, then init + apply from this directory.
#
# Use this when the repository has already been cloned (the common case here: the
# remote is a self-hosted Forgejo, so chezmoi's `init <github-user>` shorthand does
# not apply). Sourcing the repo path from the script's own location means the
# remote URL is never hardcoded.
#
#   git clone ssh://git@forgejo.home.mpli.fr/julien/dotfiles.git
#   ./dotfiles/install.sh
#
# chezmoi prompts for the machine profile on first run (see
# home/.chezmoi.toml.tmpl) and stores it in ~/.config/chezmoi/chezmoi.toml.

set -eu

if ! chezmoi="$(command -v chezmoi)"; then
    bin_dir="${HOME}/.local/bin"
    chezmoi="${bin_dir}/chezmoi"
    echo "Installing chezmoi to '${chezmoi}'" >&2
    if command -v curl >/dev/null; then
        chezmoi_install_script="$(curl -fsSL get.chezmoi.io)"
    elif command -v wget >/dev/null; then
        chezmoi_install_script="$(wget -qO- get.chezmoi.io)"
    else
        echo "To install chezmoi you must have curl or wget installed." >&2
        exit 1
    fi
    sh -c "${chezmoi_install_script}" -- -b "${bin_dir}"
    unset chezmoi_install_script bin_dir
fi

# POSIX way to get the script's own directory. --source points at the repo root;
# chezmoi reads .chezmoiroot there and takes home/ as the actual source directory.
script_dir="$(cd -P -- "$(dirname -- "$(command -v -- "$0")")" && pwd -P)"

set -- init --apply --source="${script_dir}"

echo "Running 'chezmoi $*'" >&2
exec "$chezmoi" "$@"
