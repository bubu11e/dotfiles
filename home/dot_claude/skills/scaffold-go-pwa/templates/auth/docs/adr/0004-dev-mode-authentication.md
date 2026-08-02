# 4. A development mode that relaxes authentication

Date: __DATE__

## Status

Accepted

## Context

Production sign-in needs a real email, a password, and a verified address.
During local development that is friction with no payoff: there is no mail
server, so a verification link cannot arrive, and every throwaway account needs a
password nobody will remember.

The tempting shortcut -- seeding a fixed test account, or commenting the checks
out while working -- puts a weakened auth path one forgotten commit away from
production.

## Decision

A single `auth.dev_mode` flag, false by default. When true:

- the password is optional (an empty one signs in by identity alone), though a
  supplied password is still length-checked;
- the identity field is a pseudo rather than an email, slugified into a stable
  `<pseudo>@dev.local` address so the email-keyed schema is untouched;
- accounts are created already verified, and registration signs the user in.

The server logs a warning at startup whenever the flag is on, and exposes it at
`GET /api/v1/instance` so the sign-in form can drop the password field.

## Consequences

- One flag, one code path, both modes exercised by tests. There is no separate
  "dev build" that can drift from the real one.
- The default is the safe value: omitting the setting entirely gives production
  behaviour.
- Turning the flag on in a real deployment is a serious misconfiguration, which
  is why it is a startup warning and not a debug-level line.
- Dev accounts are distinguishable forever by their `@dev.local` domain, so a
  database that was once in dev mode can be audited.
