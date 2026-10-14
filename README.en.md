<div align="center">
  <img src="docs/readme/hero.svg" width="100%" alt="KeyDesk workflow overview">
  <h1>KeyDesk</h1>
  <p><strong>Activation keys with explicit capacity and authenticated access.</strong></p>
  <p><a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-d99536?style=flat-square"></a> <a href="docs/readme/validation.md"><img alt="Validation notes" src="https://img.shields.io/badge/Validation-repository%20checks-547ac4?style=flat-square"></a></p>
  <p><a href="#overview">Overview</a> · <a href="#quick-start">Quick start</a> · <a href="#workflow">Workflow</a> · <a href="#validation">Validation</a> · <a href="#documentation">Docs</a></p>
  <p><a href="README.en.md">English</a> · <a href="README.md">简体中文</a></p>
</div>

---

<a id="overview"></a>
## Overview

A Go/PostgreSQL activation-key service and reusable package, with account authentication, key issuance and transactional activation records.

| Capability | What it provides |
| :--- | :--- |
| **Key lifecycle** | Types, expiry, capacity, validation and revocation. |
| **Authenticated API** | HS256 tokens with required expiry and administrator-only issuance. |
| **Atomic activation** | Capacity reservation and activation records share one transaction. |

<a id="quick-start"></a>
## Quick start

```bash
git clone https://github.com/eust-w/keydesk.git
cd keydesk
# Go 1.24+ and PostgreSQL are required.
go mod download
cp config/config.example.yaml config/config.yaml
# Set your database connection and a random JWT secret in config/config.yaml.
go run ./cmd/server
```

Create the PostgreSQL database named by your configuration. An administrator must explicitly grant the admin role; public registration creates ordinary users. Example role assignment:

```sql
UPDATE users SET role = 'admin' WHERE email = 'YOUR_ADMIN_EMAIL';
```

<a id="workflow"></a>
## Workflow

```text
ISSUE → AUTHORIZE → ACTIVATE
```

```bash
go test -race ./...
go build ./cmd/server
# PostgreSQL acceptance in a disposable database:
KEYDESK_TEST_POSTGRES_DSN='host=127.0.0.1 user=YOUR_USER dbname=YOUR_TEST_DB sslmode=disable' go test -race ./...
```

<a id="validation"></a>
## Validation

```bash
go test -race ./...
go build ./cmd/server
```

Recorded environment and executed checks: [validation notes](docs/readme/validation.md).

| Area | Current boundary |
| :--- | :--- |
| Migration | Legacy tokens without exp are rejected: sign in again. jwt.expires_in accepts a positive Go duration, default 24h. |
| API | Registration creates ordinary users; key generation requires an administrator role. |
| Package | Use the request-struct API in the compiled example; old positional examples are obsolete. |

<a id="documentation"></a>
## Docs

- [Compiled package example / 公共库示例](examples/basic_usage.go)
- [Activation logic / 激活逻辑](pkg/activation/manager.go)
- [Routes / 路由](cmd/server/main.go)

## License and provenance

Original code and newly created demo assets are licensed under [MIT](LICENSE). Dependencies, models, datasets and external resources retain their terms; see [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
