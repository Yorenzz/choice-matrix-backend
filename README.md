# choice-matrix-backend

## Run

### Normal start

```powershell
go run .
```

### Hot reload with Air

```powershell
air
```

If `air` is not available in your terminal PATH yet, run it with the full path:

```powershell
C:\Users\Yorenz\go\bin\air.exe
```

Project config file:

- `.air.toml`

## Auth and Redis

The backend now uses:

- `ACCESS_TOKEN_TTL` for short-lived access tokens, default `15m`
- `REFRESH_TOKEN_TTL` for Redis-backed refresh sessions, default `168h`
- `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` / `REDIS_DB`
- `SKIP_AUTO_MIGRATE` to skip `AutoMigrate` while still connecting to PostgreSQL

Both PostgreSQL and Redis need to be available before the server starts.
