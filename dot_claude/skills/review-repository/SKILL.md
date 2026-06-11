---
name: review-repository
description: To perform a comprehensive repository health check, ensure security compliance, and align the project with "State-of-the-Art" standards.
---

## Execution Protocol

1. **Detect Tech Stack**
   - Identify the primary language and framework by checking for: `Pipfile` / `pyproject.toml` (Python), `package.json` (Node), `go.mod` (Go), `Cargo.toml` (Rust), `pom.xml` / `build.gradle` (JVM), `*.yml` playbooks (Ansible), etc.
   - Use this to inform all subsequent steps.

2. **Initialize `CLAUDE.md`**
   - Check if `CLAUDE.md` exists at the repo root.
   - If missing: create it, populated with the detected build, test, and lint commands.
   - If present: verify it reflects current commands. Flag outdated entries as `High (Setup)` in `TODO.md` — do not modify it automatically.

3. **Security & Gitleaks**
   - Check for `.pre-commit-config.yaml`.
   - If missing: add a `Critical (Security)` task to `TODO.md` to create it with `gitleaks`.
   - If present but `gitleaks` hook is absent: add a `Critical (Security)` task to add it.
   - If `gitleaks` is already configured: note it as compliant, no action needed.

4. **Git Hygiene**
   - Audit `.gitignore` for missing standard exclusions based on the detected stack:
     - OS: `.DS_Store`, `Thumbs.db`
     - IDE: `.idea/`, `.vscode/`, `*.swp`
     - Build artifacts: `dist/`, `build/`, `__pycache__/`, `*.pyc`, `node_modules/`, `target/`
     - Secrets: `.env`, `*.enc`, `*.key`, `*.pem`
   - Flag each missing category as `Medium (Maintenance)`.

5. **State-of-the-Art Audit**
   - Check for the presence of these standard config files (flag missing ones as `High (Setup)`):
     - Linter config: `.eslintrc`, `pyproject.toml [tool.ruff]`, `.golangci.yml`, etc.
     - Formatter config: `.prettierrc`, `pyproject.toml [tool.black]`, etc.
     - CI config: `.github/workflows/`, `.woodpecker.yml`, `.gitlab-ci.yml`, etc.
     - Dependency lockfile: `Pipfile.lock`, `package-lock.json`, `go.sum`, etc.
     - `.editorconfig` for cross-editor consistency.

6. **Test Gap Analysis**
   - List all source files containing business logic.
   - For each, check whether a corresponding test file exists (e.g., `src/foo.py` -> `tests/test_foo.py`).
   - Flag missing test files as `Medium (Maintenance)` in `TODO.md`.

## TODO.md Template

```markdown
# Repository Health TODO

**Date:** <date>
**Stack:** <detected language/framework>

## Critical (Security)

- [ ] <task description> -- <reason>

## High (Setup)

- [ ] <task description> -- <reason>

## Medium (Maintenance)

- [ ] <task description> -- <reason>
```

Add one entry per finding. Omit a section if it has no entries.

## Output Rules

- **Non-Destructive**: Do NOT modify source code, install packages, or alter git history.
- **CLAUDE.md exception**: The only file that may be created (not modified) is `CLAUDE.md`, and only if it is entirely absent.
- **Reporting**: All findings must be written to `TODO.md` at the repo root using the template above.
