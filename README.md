# Chirp API

Chirp is a Go-powered JSON API for a microblogging platform. It exposes endpoints for account management, authentication, and publishing short "chirps" while serving a static frontend from the `public/` directory. PostgreSQL stores all persistent data and `sqlc` generates the typed query layer that backs the handlers.

## Features

- JWT-based authentication with refresh tokens and password hashing.
- Chirp CRUD endpoints with author filtering, sort controls, and ownership checks on deletion.
- User profile management, including credential updates and a "Chirpy Red" upgrade webhook.
- Environment-aware admin utilities such as health checks, metrics, and a development-only reset endpoint.
- Static asset hosting via Go's `http.FileServer` behind instrumentation middleware.

## Tech Stack

- Go 1.24
- PostgreSQL 14+
- `sqlc` for query generation (`internal/database`)
- Goose-style SQL migrations (`sql/schema`)
- `github.com/golang-jwt/jwt/v5` for token handling
- `golang.org/x/crypto` for BCrypt password hashing

## Getting Started

1. **Install prerequisites**
   - Go 1.24 or newer
   - PostgreSQL running locally or accessible via connection string
   - Optional tooling: `sqlc` and `goose` (or your preferred migration runner)

2. **Clone the repository**
   ```bash
   git clone https://github.com/Sanghun1Adam1Park/chirp.git
   cd chirp
   ```

3. **Configure environment variables**
   Create a `.env` file (automatically loaded via `godotenv`) or export the variables in your shell.
   ```env
   DB_URL=postgres://user:pass@localhost:5432/chirp?sslmode=disable
   PLATFORM=dev                # enables /admin/reset when set to "dev"
   SECRET=your-jwt-signing-key
   POLKA_KEY=shared-secret-for-polka-webhooks
   ```

4. **Run database migrations**
   ```bash
   goose -dir sql/schema postgres "$DB_URL" up
   ```
   Adjust the command if you use a different migration tool; all migrations live in `sql/schema`.

5. **Generate query code (optional, when queries change)**
   ```bash
   sqlc generate
   ```

6. **Start the API server**
   ```bash
   go run .
   ```
   The server listens on `http://localhost:8080`. Static assets are served from `/app/*`.

## API Overview

All JSON endpoints respond with `application/json`. Authentication endpoints issue JWT access tokens; protected routes expect `Authorization: Bearer <token>` headers.

- `GET /api/healthz` — Plaintext readiness probe.
- `GET /admin/metrics` — HTML stats page showing static file hits.
- `POST /admin/reset` — Development-only helper that truncates user data when `PLATFORM=dev`.
- `POST /api/users` — Register a user with `email` and `password`.
- `POST /api/login` — Authenticate and receive access plus refresh tokens.
- `PUT /api/users` — Update email and password for the authenticated user.
- `POST /api/refresh` — Exchange a refresh token (sent in the `Authorization` header) for a new access token.
- `POST /api/revoke` — Revoke the provided refresh token.
- `POST /api/chirps` — Create a chirp for the authenticated user.
- `GET /api/chirps` — List chirps.
  - Optional query params: `author_id=<uuid>` filters to an author's posts; `sort=asc|desc` controls chronological order (`asc` default).
- `GET /api/chirps/{id}` — Fetch a single chirp by ID.
- `DELETE /api/chirps/{chirpID}` — Delete a chirp you own.
- `POST /api/polka/webhooks` — Accept Polka webhook events (expects `Authorization: Bearer <POLKA_KEY>`); processes `user.upgraded` to toggle the `is_chirpy_red` flag.

## Database Schema

Migrations live in `sql/schema/` and create the following core tables:

- `users` — Stores account metadata, hashed passwords, and the `is_chirpy_red` flag.
- `chirps` — Contains short-form posts linked to users.
- `refresh_tokens` — Tracks refresh tokens, expiry, and revocation timestamps.

Corresponding query definitions are in `sql/queries/`; running `sqlc generate` regenerates the Go client in `internal/database/`.

## Development

- Run tests: `go test ./...`
- Format code: `gofmt -w <files>`
- Regenerate SQL bindings after query or schema changes: `sqlc generate`

Feel free to extend the API, add validation, or wire in a frontend—Chirp is intended as a solid foundation for experimenting with Go HTTP services.
