# Grinch 🎄

Grinch is a small Go + React control plane for Santa. It keeps users and groups in sync from Entra ID, hands rules to Santa clients, collects execution and file access events, and gives admins a simple UI to manage it all.

We created Grinch as other sync servers are quite basic in terms of who/which machines get what.
It works for us, it is not fancy, but it will get better over time, [if syncv2 is ungated...](https://github.com/northpolesec/santa/blob/da9f1fe4555f52823254c8fa949f9b3e18f6563b/Source/common/Pinning.mm#L22-L28)

> [!WARNING]
> This project may be unstable or have bugs, use with caution.
> Also expect breaking changes between releases for now.

## ✨ Features

- Santa sync endpoints (`/sync`) for preflight, rule download, event upload, and postflight
- React Admin UI for rules, machines, executables, execution events, file access events, users, and groups
- Entra ID sync for users, groups, and memberships
- Local admin login or Entra OAuth
- Postgres database
- Container-first deployment that can serve the built frontend from the backend

## 🧭 How it fits together

- The backend serves `/api/v1`, `/auth` (login/cookies), `/sync` (Santa clients), and `/` when the frontend build is present.
- The frontend is a React Admin app that talks to `/api/v1`.
- Rules define what Santa should allow, block, silent block, or evaluate with CEL.
- Groups provide the targeting layer. Rules are attached to groups with include/exclude scopes and priority.
- Santa clients talk to `/sync` and receive rule updates plus event ingestion.

## 🚀 Deploy (Docker)

1. Create a `.env` file and fill in the required values below.
2. Start the stack:

```bash
docker compose up --build
```

The app listens on `http://localhost:8080`.
The Docker image builds the frontend and serves it from the backend automatically.

## 🧰 Configuration

| Name                       | What it does                               | Required                  | Notes                                                      |
| -------------------------- | ------------------------------------------ | ------------------------- | ---------------------------------------------------------- |
| `GRINCH_PORT`              | HTTP listen port                           | No                        | Defaults to `8080`.                                        |
| `GRINCH_BASE_URL`          | Public URL for cookies and OAuth           | Yes, when auth is enabled | Must be the externally reachable URL.                      |
| `LOG_LEVEL`                | Log verbosity                              | No                        | `debug`, `info`, `warn`, `error`.                          |
| `DATABASE_HOST`            | Postgres host                              | Yes                       |                                                            |
| `DATABASE_PORT`            | Postgres port                              | No                        | Defaults to `5432`.                                        |
| `DATABASE_USER`            | Postgres user                              | Yes                       |                                                            |
| `DATABASE_PASSWORD`        | Postgres password                          | Yes                       |                                                            |
| `DATABASE_NAME`            | Postgres database name                     | Yes                       |                                                            |
| `DATABASE_SSLMODE`         | Postgres SSL mode                          | No                        | Defaults to `disable`.                                     |
| `JWT_SECRET`               | Signing secret for auth                    | Yes, when auth is enabled | Keep it dedicated to JWT signing.                          |
| `LOCAL_ADMIN_PASSWORD`     | Enable local admin login                   | No                        | Username is always `admin`.                                |
| `ENTRA_TENANT_ID`          | Entra tenant ID                            | No                        | Set with the other `ENTRA_*` vars for Entra auth and sync. |
| `ENTRA_CLIENT_ID`          | Entra client ID                            | No                        | Set with the other `ENTRA_*` vars for Entra auth and sync. |
| `ENTRA_CLIENT_SECRET`      | Entra client secret                        | No                        | Set with the other `ENTRA_*` vars for Entra auth and sync. |
| `ENTRA_SYNC_ENABLED`       | Enable periodic Entra sync                 | No                        | Defaults to `false`.                                       |
| `ENTRA_SYNC_INTERVAL`      | Entra sync interval                        | No                        | Defaults to `1h` when enabled.                             |
| `EVENT_RETENTION_DAYS`     | How long to keep stored events             | No                        | Defaults to `90`.                                          |
| `EVENT_DECISION_ALLOWLIST` | Optional decision filter for stored events | No                        | Comma-separated decision names.                            |

## 🖥️ Santa client setup

Grinch speaks the Santa sync protocol at:

```
https://grinch.awesomeit.net/sync
```

Set these keys in your Santa configuration profile:

```xml
<key>SyncBaseURL</key>
<string>https://grinch.awesomeit.net/sync</string>
<key>SyncEnableProtoTransfer</key>
<true/>
<key>MachineOwner</key>
<string>{{userprincipalname}}</string>
<key>SyncClientContentEncoding</key>
<string>gzip</string>
```

`MachineOwner` is optional, but if you use it, it should be the user's UPN/email.
Grinch uses it for primary-user matching and user-group targeting.

## 🧾 Rules and targeting

Rules:

- A rule is a reusable template (binary hash, signing ID, team ID, certificate, cdhash, etc).
- Each rule has a type and an identifier.
- Rules can include custom message/URL metadata.

Targeting:

- A rule has attachments that target groups.
- Each attachment has include groups, optional exclude groups, a priority, and the Santa policy to apply.
- Evaluation is deterministic: attachments are checked in priority order and the first matching include wins.
- A machine’s effective groups come from direct machine group membership plus primary-user membership.
- The server sends at most one effective Santa rule per `(rule_type, identifier)`.

Typical flow:

1. Create rules.
2. Create groups for the users or machines you want to target.
3. Attach rules to those groups with the policy you want.
4. Santa clients sync and receive only the effective rules for that machine.

## 🧾 Executables and events

- `executables` are first-class records for observed binaries/processes.
- `execution-events` record execution decisions on machines.
- `file-access-events` record file access decisions and process chains.
- Raw events are not reconstructable as they come in on the wire

## 🧪 Local development

Backend:

```bash
cd backend
go mod download
go generate ./...
go run ./cmd/grinch
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

Vite proxies `/api` and `/auth` to `localhost:8080`.
The backend serves the built frontend in container deployments.

## ⚠️ Limitations

- No RBAC; anyone who can log in is an admin.
- Entra is the only directory sync source implemented.
- No rate limiting.

## 🤝 Contributing / PRs

We are happy to take PRs. Fork this repo, make your changes, and open a PR.
Feel free to open an [issue](https://github.com/woodleighschool/grinch) if you find any bugs or to request a feature.
