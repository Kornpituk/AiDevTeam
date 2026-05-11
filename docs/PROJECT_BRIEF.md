# Project Brief: AiDevT

## Product Name

AiDevT - AI Development Task Orchestrator

## Current Goal

**Agent Orchestration Engine + Dashboard for AI-assisted software development.**

Build a system that allows:
1. Users to define reusable agent profiles (identity, role, system prompt, model preferences)
2. Users to define agent teams (collections of profiles with roles and execution order)
3. Users to start agent runs for specific development tasks
4. Backend to automatically execute ordered steps using an orchestration engine
5. Dashboard to track run status, steps, messages, approvals, tool calls, errors, and summaries

## Non-Goals (for now)

These features are explicitly out of scope for the current MVP:

1. **No Authentication / Authorization**: No user accounts, login, OAuth, JWT, or permission systems.
2. **No Distributed Queue**: No Redis, RabbitMQ, or external job queue. Execution will be in-memory or simple database polling.
3. **No WebSocket Requirement**: No real-time push updates. Dashboard will use polling if needed.
4. **No Automatic Codebase Modification by AI**: All AI execution must be explicitly planned, scoped, and approved before any code changes.
5. **No Secrets in Code**: No hardcoded API keys, tokens, or credentials. All config must come from environment variables.
6. **No Multi-tenant**: Single workspace/project focus for now.

## MVP Success Criteria

The MVP is complete when:

1. ✅ **User can create task** via UI or API
2. ✅ **User can create agent profiles/teams** via UI
3. ✅ **User can start an agent run** (POST /agent-runs/:id/start)
4. ✅ **Backend can auto-create and execute ordered run steps** from team members when appropriate
5. ✅ **Step statuses transition correctly** (pending → running → completed/failed/cancelled)
6. ✅ **Run status transitions correctly** (draft → running → completed/failed/cancelled)
7. ✅ **Messages/output/error are persisted** to database
8. ✅ **Dashboard updates via polling** to show current execution state
9. ✅ **No secrets are hardcoded** in any file
10. ✅ **All tests pass** (backend: `go test ./...`, frontend: `npm run test:run`, `npm run lint`, `npm run build`)

## Stack

- **Frontend**: Next.js 14 + TypeScript in `apps/web`
- **Backend**: Go 1.22+ in `services/api`
- **Database**: Postgres 16+
- **Migrations**: SQL files in `db/migrations`
- **Container**: Docker + docker-compose for local development

## Development Principles

1. **Small patches**: Keep each change focused and reviewable
2. **Plan before edit**: Always read relevant docs and understand the scope before modifying code
3. **Test after change**: Backend: `go test ./...`, Frontend: `npm run test:run`, `npm run lint`, `npm run build`
4. **Show diff**: After making changes, show git diff for review
5. **No old migration changes**: Never modify existing migration files after they are applied
6. **No .env edits**: Do not modify .env files; use .env.example as reference
7. **Phase discipline**: Work on one phase at a time; do not skip ahead
