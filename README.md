# AiDevT

**Agent Orchestration Engine + Dashboard for AI-assisted software development.**

## Quick Start

### Prerequisites

- Docker / Docker Compose
- Go 1.22+
- Node.js 18+
- npm

### 1. Configure Environment

Copy the example environment file and edit values as needed:

```bash
cp .env.example .env
```

For local demo mode, keep:

```env
LLM_PROVIDER=fake
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

Use `LLM_PROVIDER=fake` when you want to run the project without an external LLM API key.

### 2. Start Postgres

From the repository root:

```bash
docker-compose up -d db
```

### 3. Apply Migrations

From the repository root:

```bash
docker exec -i aidevt-db psql -U postgres -d aidevt < db/migrations/000001_init.sql
docker exec -i aidevt-db psql -U postgres -d aidevt < db/migrations/000002_agent_orchestration.sql
docker exec -i aidevt-db psql -U postgres -d aidevt < db/migrations/000003_approval_tool_link.sql
```

### 4. Run Backend API

The backend loads environment variables from `.env` in either:

- the current directory (`services/api/.env`), or
- the project root (`../../.env` when running from `services/api`)

For local demo mode, make sure the project root `.env` includes:

```env
LLM_PROVIDER=fake
```

Use a real provider only when you have an API key configured:

```env
LLM_PROVIDER=openai
LLM_API_KEY=your-api-key
LLM_MODEL=gpt-4
```

Open a new terminal:

```bash
cd services/api
go run cmd/api/main.go
```

Backend API runs at:

```text
http://localhost:8080
```

Health check:

```bash
curl http://localhost:8080/health
```

### 5. Run Frontend

Open another terminal:

```bash
cd apps/web
npm install
npm run dev
```

Frontend runs at:

```text
http://localhost:3000
```

If port `3000` is busy, Next.js may use `3001`.

### 6. Open the App

Start here:

```text
http://localhost:3000/tasks
```

Useful pages:

- `http://localhost:3000/tasks` - Task list
- `http://localhost:3000/tasks/new` - Create task
- `http://localhost:3000/agents` - Agent profiles and teams
- `http://localhost:3000/agent-runs/[id]` - Agent run detail

## Development Commands

### Backend

```bash
cd services/api
go test ./...
go run cmd/api/main.go
```

### Frontend

```bash
cd apps/web
npm run test:run
npm run lint
npm run build
npm run dev
```

## Database Commands

Connect to database:

```bash
psql postgresql://postgres:postgres@localhost:5432/aidevt
```

Stop database:

```bash
docker-compose down
```

Stop and delete volume (destroys all local data):

```bash
docker-compose down -v
```

## Connection String

```
postgresql://postgres:postgres@localhost:5432/aidevt
```

## Environment Variables

Common local variables:

```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/aidevt
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080

LLM_PROVIDER=fake
LLM_API_KEY=
LLM_MODEL=gpt-4
LLM_BASE_URL=
LLM_TIMEOUT=60s

WORKSPACE_ROOT=
TOOL_READ_MAX_BYTES=1048576
TOOL_SEARCH_MAX_RESULTS=50
TOOL_MAX_ITERATIONS=10
```

Do not commit `.env` files or real API keys.

## Project Docs

For full documentation, see:

- [docs/PROJECT_BRIEF.md](docs/PROJECT_BRIEF.md) - Product vision, goals, non-goals
- [docs/CURRENT_STATUS.md](docs/CURRENT_STATUS.md) - What exists, what doesn't, API reference
- [docs/DEMO_RUNBOOK.md](docs/DEMO_RUNBOOK.md) - Step-by-step demo guide with setup instructions
- [docs/PHASE_C_PLAN.md](docs/PHASE_C_PLAN.md) - Detailed plan for auto orchestration
- [docs/AGENT_TEAM.md](docs/AGENT_TEAM.md) - Agent definitions, ownership, coordination
- [docs/WORK_LOG.md](docs/WORK_LOG.md) - History, current status, next steps
