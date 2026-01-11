# Grinch 🎄

Grinch is a small Go + React control plane for Santa. It keeps users and groups in sync from Entra ID, hands rules to Santa clients, collects decision logs, and gives admins a simple UI to manage it all.

We created Grinch as other sync servers are quite basic in terms of who/which machines get what.
It works for us, it is not fancy, but it will get better over time.

## ✨ Features

-   Santa sync endpoints (`/sync`) for preflight, rule download, event upload, and postflight
-   React Admin UI for rules, policies, machines, events, users, and groups
-   Entra ID sync for users, groups, and memberships
-   Local admin login or Microsoft OAuth
-   Postgres database

## 🧭 How it fits together

-   The backend serves `/api` (admin UI), `/auth` (login/cookies), `/sync` (Santa clients), and `/` when the frontend build is present.
-   The frontend is a React Admin app that talks to `/api`.
-   Policies define client settings and which rules apply to which targets.
-   Santa clients talk to `/sync` and receive policy settings and rule sets.

## 🚀 Deploy (Docker)

1. Copy `.env.example` to `.env` and fill in the required values.
2. Start the stack:

```bash
docker compose up --build
```

The app listens on `http://localhost:8080` and serves the frontend when `FRONTEND_DIR` is set (the Dockerfile does this for you).

### Production notes

-   ⚠️ Put it behind HTTPS (Caddy, Nginx, whatever you like). ⚠️
-   Set `BASE_URL` to the public URL (used for auth cookies and OAuth callbacks).
-   Set a strong `AUTH_SECRET` (minimum 32 chars).
-   Keep Postgres data on a volume.

## 🧰 Configuration

| Name                      | What it does                     | Required | Notes                                      |
| ------------------------- | -------------------------------- | -------- | ------------------------------------------ |
| `PORT`                    | HTTP listen port                 | No       | Defaults to `8080`.                        |
| `LOG_LEVEL`               | Log verbosity                    | No       | `debug`, `info`, `warn`, `error`.          |
| `BASE_URL`                | Public URL for cookies and OAuth | Yes      | Must be the externally reachable URL.      |
| `FRONTEND_DIR`            | Path to built frontend           | No       | Used when serving the UI from the backend. |
| `DB_HOST`                 | Postgres host                    | Yes      |                                            |
| `DB_PORT`                 | Postgres port                    | No       | Defaults to `5432`.                        |
| `DB_USER`                 | Postgres user                    | Yes      |                                            |
| `DB_PASSWORD`             | Postgres password                | Yes      |                                            |
| `DB_NAME`                 | Postgres database name           | Yes      |                                            |
| `DB_SSLMODE`              | Postgres SSL mode                | No       | Defaults to `disable`.                     |
| `AUTH_SECRET`             | Signing secret for auth          | Yes      | Must be at least 32 chars.                 |
| `TOKEN_DURATION`          | Auth token lifetime              | No       | Defaults to `1h`.                          |
| `COOKIE_DURATION`         | Auth cookie lifetime             | No       | Defaults to `24h`.                         |
| `ADMIN_PASSWORD`          | Enable local admin login         | No       | Username is always `admin`.                |
| `MICROSOFT_CLIENT_ID`     | Microsoft OAuth client ID        | No       | Enable Microsoft login.                    |
| `MICROSOFT_CLIENT_SECRET` | Microsoft OAuth client secret    | No       | Enable Microsoft login.                    |
| `ENTRA_TENANT_ID`         | Entra tenant ID                  | Yes      | Required for Entra sync.                   |
| `ENTRA_CLIENT_ID`         | Entra client ID                  | Yes      | Required for Entra sync.                   |
| `ENTRA_CLIENT_SECRET`     | Entra client secret              | Yes      | Required for Entra sync.                   |
| `ENTRA_SYNC_INTERVAL`     | Entra sync interval              | No       | Defaults to `15m`.                         |

See `.env.example` for the full list.

## 🖥️ Santa client setup

Grinch speaks the standard Santa sync protocol at:

```
https://grinch.awesomeit.net/sync
```

Set these keys in your Santa configuration profile:

```xml
<key>SyncBaseURL</key>
<string>https://grinch.awesomeit.net/sync</string>
<key>MachineOwner</key>
<string>{{userprincipalname}}</string>
<key>SyncClientContentEncoding</key>
<string>gzip</string>
```

It is expected that `MachineOwner` machines the UPN of the user.

## 🧾 Rules and policies

Rules:

-   A rule is a reusable template (binary hash, signing ID, team ID, etc).
-   Each rule has a type and an identifier.
-   Rules can include custom message/URL metadata.

Policies:

-   A policy defines the Santa client settings (mode, batch size, regex overrides, etc).
-   Policies target users, groups, machines, or all.
-   You attach rules to policies with an action (allow, block, silent block, or CEL).
-   Priority matters: higher numbers win. A machine only ever ends up with one effective policy.
-   Evaluation picks the highest-priority policy that matches the machine/user/group targets.

Typical flow:

1. Create rules (allow/block templates).
2. Create a policy for a group or machine set.
3. Attach the rules to the policy.
4. Santa clients sync and apply the policy.

Due to how policies work, it's really only designed to have 1 policy assigned to `All Machines` with a priority of `0` (to make a base policy).

## 🧪 Local development

Backend:

```bash
cd backend
go mod download
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
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

## ⚠️ Limitations

-   No auth on `/sync` (yet).
-   Only Entra ID sync is implemented.
-   No RBAC; anyone who can log in is an admin.
-   CANNOT be horizontally scaled (yet)

## 🤝 Contributing / PRs

We are happy to take PRs. Fork this repo, make your changes, and open a PR.
Feel free to open an [issue](https://github.com/woodleighschool/grinch) if you find any bugs or to request a feature!
