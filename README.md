Status page for [PluralKit](https://github.com/PluralKit/PluralKit) built with SvelteKit and Go.

This application is designed to be run behind a proxy of some kind (see provided Caddyfile), with the static frontend files served by the proxy (alternatively, frontend files could be served using any http server, such as GitHub Pages or Cloudflare Pages).

The frontend is seperated from the backend, pulling status information from a publicly accessible API url, without which it will still function as a shard status page (provided the shard status url is correctly set).

Incidents can be created using the HTTP api with an auth token (currently simple auth specified by an environment variable). Data is stored in an SQLite3 file `status.db`, but can theoretically be updated to use other databases supported by [Bun](https://bun.uptrace.dev/).

See [TODO: routes.md](./routes.md) for more details on the backend API routes.

## Development
Requirements:
- `Go 1.24`
- `node.js v20`
- `make`

The frontend is built in node.js/SvelteKit with prerendering enabled, and backend built in Go.

A makefile is provided for convience, simply run `make` to build the frontend/backend, or `make dev` to build the frontend/backend and run in development mode (static files are served using the Chi router).

## Running
A docker compose file is provided for easy testing, simply create a `.env` file specifying the following:
```
AUTH_TOKEN=
NOTIFICATION_WEBHOOK=
NOTIFICATION_ROLE=
```
and run `docker compose up -d`

## Environment Variables
TODO: add details

### Backend:
``` golang
type Config struct {
	BindAddr            string    `env:"pluralkit__status__addr" envDefault:"0.0.0.0:8080"`
	AuthToken           string    `env:"pluralkit__status__auth_token"`
	NotificationWebhook string    `env:"pluralkit__status__notification_webhook"`
	NotificationRole    string    `env:"pluralkit__status__notification_role"`
	RunDev              bool      `env:"pluralkit__status__run_dev" envDefault:"false"`
	DBLoc               string    `env:"pluralkit__status__db_location" envDefault:"file:status.db"`
	LogLevel            SlogLevel `env:"pluralkit__consoleloglevel" envDefault:"info"`
}
```

### Frontend:
```
```