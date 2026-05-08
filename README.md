# AiDevT

AI Dev Task Dashboard

## Quick Start

### Database Setup

1. Start Postgres:
```bash
docker-compose up -d
```

2. Apply initial migration:
```bash
docker exec -i aidevt-db psql -U postgres -d aidevt < db/migrations/000001_init.sql
```

3. Connect to database:
```bash
psql postgresql://postgres:postgres@localhost:5432/aidevt
```

### Stop Database

```bash
docker-compose down
```

To stop and delete volume (destroy all data):
```bash
docker-compose down -v
```

## Connection String

```
postgresql://postgres:postgres@localhost:5432/aidevt
```

Copy `.env.example` to `.env` if needed for local configuration.
