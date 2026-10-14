# Release validation

Date: 2026-10-14. Checks are executed in an isolated working checkout. Screenshots use local synthetic/demo data. A source implementation, a passing automated test and external-service acceptance are separate claims.

## Commands

```bash
go test -race ./...
go build ./cmd/server
```

PASS: Go race tests/build plus PostgreSQL HTTP permissions, invalid-length rejection, expired/legacy/wrong-algorithm tokens, repeated activation and a 12-client race for one slot. The race produced one successful activation and one record. Hardware, external credentials, signing and production deployment are separate setup/acceptance steps.
