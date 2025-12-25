# Contributing to bssbackendgo

Thanks for contributing to **bssbackendgo** 🎮  
This project enforces a strict identity and authorization architecture to
prevent security and data integrity regressions.

Please read this before making changes.

---

## Core Rule

**Architecture is enforced, not advisory.**

If a change violates documented invariants, CI will fail.
This is intentional.

---

## Before You Make Changes

If your work touches any of the following:

- Authentication or session handling
- Authorization logic
- Identity fields
- Middleware
- Players / users tables
- Ownership checks
- Context values

You **must** review the architecture documentation first.

### Required Reading

- [`docs/architecture.md`](docs/architecture.md)  
  → System architecture, trust boundaries, and invariants

- `README.md` (Architecture Overview section)  
  → High-level identity and request flow

---

## Architecture Enforcement Scripts

This repository includes automated checks that enforce architectural rules.

Location:

```text
/scripts/architecture/
```

These scripts are:

Run in CI

Mapped directly to sections in docs/architecture.md

Treated as executable documentation

If a script fails, it usually means an important invariant was violated.

See:

Architecture enforcement overview (internal doc)

Script names match architecture sections

Making Architectural Changes

If you intentionally change architecture behavior, you must update both:

docs/architecture.md

The corresponding script in scripts/architecture/

Architecture drift is considered a bug.

What Not To Do

Do not trust client-provided IDs

Do not parse JWTs in handlers

Do not use external auth IDs as database primary keys

Do not enforce authorization in frontend code

Do not bypass sessionValidationMiddleware

If you are unsure whether a change is safe, open a PR and ask.

Style & Expectations

Keep handlers simple

Enforce authorization in SQL

Prefer clarity over cleverness

Favor explicit invariants over implicit behavior

Thanks for helping keep the system safe, auditable, and maintainable.

## Pull Request Checklist

Before opening a PR, confirm:

- [ ] I reviewed `docs/architecture.md`
- [ ] I did not trust client-provided identity data
- [ ] All authorization checks use internal IDs (`players.id`)
- [ ] Handlers do not parse JWTs or call Descope directly
- [ ] Architectural changes update both docs **and** enforcement scripts
- [ ] Architecture enforcement scripts pass

If any item is unchecked, explain why in the PR description.
