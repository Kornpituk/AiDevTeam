# AiDevT

**Agent Orchestration Engine + Dashboard for AI-assisted software development.**

## Quick Start

### Database Setup

1. Start Postgres:
```bash
docker-compose up -d
```

2. Apply initial migration:
```bash
docker exec -i aidevt-db psql -U postgres -d aidevt < db/migrations/000001_init.sql
docker exec -i aidevt-db psql -U postgres -d aidevt < db/migrations/000002_agent_orchestration.sql
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

## Project Docs

For full documentation, see:

- [docs/PROJECT_BRIEF.md](docs/PROJECT_BRIEF.md) - Product vision, goals, non-goals
- [docs/CURRENT_STATUS.md](docs/CURRENT_STATUS.md) - What exists, what doesn't, API reference
- [docs/DEMO_RUNBOOK.md](docs/DEMO_RUNBOOK.md) - Step-by-step demo guide with setup instructions
- [docs/PHASE_C_PLAN.md](docs/PHASE_C_PLAN.md) - Detailed plan for auto orchestration
- [docs/AGENT_TEAM.md](docs/AGENT_TEAM.md) - Agent definitions, ownership, coordination
- [docs/WORK_LOG.md](docs/WORK_LOG.md) - History, current status, next steps
