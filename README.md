# Chirpy

🐦 This is a guided project from [Boot.dev](https://boot.dev). It teaches you how to build an API in Golang including but not limited to:

- CRUD operations with PostgreSQL
- Webhooks
- Authentication and Authorization
- Routing
- Adding query parameters as options for HTTP requests

# How To Use

1. Copy `.env.example` to `.env` file.
2. Populate the environmental variables.

## How To Populate your envs

Generate a `SIGNING_KEY` key by running the following command.

```bash
openssl rand -base64 64
```

Copy the value and paste it between the double-quotes e.g.

```
SIGNING_KEY="thisIsMyKEY"
```

### Setup PostgreSQL

Start the Postgres server in the background
  - Mac: brew services start postgresql@15
  - Linux: sudo service postgresql start

Connect to the server. I recommend simply using the psql client. It's the "default" client for Postgres, and it's a great way to interact with the database. While it's not as user-friendly as a GUI like PGAdmin, it's a great tool to be able to do at least basic operations with.

Enter the psql shell:
  - Mac: `psql postgres`
  - Linux: `sudo -u postgres psql`

You should see a new prompt that looks like this:

```
postgres=#
```

Create a new database.

```
CREATE DATABASE chirpy;
```

Connect to the new database:

```
\c chirpy
```

You should see a new prompt that looks like this:

```
chirpy=#
```

Set the user password (Linux only)

```
ALTER USER postgres PASSWORD 'postgres';
```

You can type `exit` to leave the `psql` shell.

Get your connection string. A connection string is just a URL with all of
the information needed to connect to a database. The format is:

```
protocol://username:password@host:port/database
```

Paste the connection string to `DB_URL` e.g.

```
DB_URL="postgres://postgres:@localhost:5432/chirpy?sslmode=disable"
```

Make sure that you have the `?sslmod=disable` as the last part of the string.

### Join [boot.dev](https://www.boot.dev/)

Head to the Go backend pathway. You have to reach to **Learn HTTP Servers in GO** since
you have to get your `POLKA_KEY`.

### Setup the migrations

Install goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Install `just`. See <https://github.com/casey/just>.

Then run the following command.

```bash
just up-migrate
```

This should prepare all the tables needed for this project to work.

### Run it

```
just run
```

This will build and run the command `./chirpy`.

### Experiment with it

Here are the available API endpoints:

- `/app/` -> This just opens up a page to [index.html](./index.html).
- `POST /api/chirps` -> pass a JSON object with this shape: `{"body": "body string" }`. You need to be authorized to call this endpoint though so get your token and prepare an Authorization header with this format `Bearer <token>`.
- `GET /api/chirps` -> Gets all chirps. You can pass `author_id` e.g. `chirps?author_id=ID` here. You can also pass `sort` as well e.g. `chirps?sort=asc` or `chirps?sort=desc`.
- `GET /api/chirps/{chirpID}` Get a chirp based by chirp ID.
- `DELETE /api/chirps/{chirpID}` Delete a chirp by chirp ID. Requires authorization. You need to be authorized to call this endpoint though so get your token and prepare an Authorization header with this format `Bearer <token>`.
- `GET /api/healthz`
- `POST /api/users` -> Register your user here. Just pass a shape `{"email": "email@email.com", "password": "strong password"}`.
- `PUT /api/users` -> Update your user here. Just pass a shape `{"email": "email@email.com", "password": "strong password"}`.
- `POST /api/login` -> You will get your token here. Just pass a shape like `{"email": "email@email.com", "password": "strong password"}`. You have to register first.
- `POST /api/revoke` -> You need to be authorized to call this endpoint.
- `POST /api/refresh` -> You need to be authorized to call this endpoint by passing a Bearer token where token is your **refresh** token.
- `POST /api/polka/webhooks` -> You need to pass a shape `{"event": "kind", "data": { "moredata": "moredata" }}`.
- `GET /admin/metrics`
- `POST /admin/reset` -> You need to be authorized to call this endpoint so get your token and prepare an Authorization header with this format `Bearer <token>`.
