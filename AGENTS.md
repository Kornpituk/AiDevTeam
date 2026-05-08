# AiDevT

AiDevT is an AI Dev Task Dashboard.

## Goal

Build a control plane for AI-assisted software development.

The first MVP should allow users to:
- Create development tasks
- Store task plans
- Store review notes
- Track task status
- View task events and artifacts

## Stack

- Frontend: Next.js + TypeScript in `apps/web`
- Backend: Go in `services/api`
- Database: Postgres
- Migrations: SQL files in `db/migrations`

## Project Structure

- `apps/web`: Next.js frontend
- `services/api`: Go backend API
- `db/migrations`: SQL migrations
- `docs`: architecture and workflow docs

## Development Rules

- Always plan before editing files.
- Keep patches small.
- Do not implement multiple phases at once.
- Do not edit `.env` files.
- Do not add auth yet.
- Do not add AI automation yet.
- Do not create destructive database migrations.
- Do not modify old migrations after they are applied.
- After code changes, show `git diff`.
- Backend changes should run `go test ./...`.
- Frontend changes should run `npm run lint` and `npm run build` when available.

## MVP Phases

### Phase 1: Database foundation

Create:
- `docker-compose.yml`
- Postgres service
- initial SQL migration for tasks, events, and artifacts
- README instructions

Database schema should include:
- `ai_tasks`
- `ai_task_events`
- `ai_task_artifacts`

### Phase 2: Go API

Create:
- health endpoint
- task create/list/detail
- task status update
- task events
- task artifacts

Required endpoints:
- GET /health
- POST /tasks
- GET /tasks
- GET /tasks/:id
- PATCH /tasks/:id/status
- POST /tasks/:id/events
- GET /tasks/:id/events
- POST /tasks/:id/artifacts
- GET /tasks/:id/artifacts

### Phase 3: Next.js dashboard

Create:
- task list page
- create task form
- task detail page
- event timeline
- artifact viewer

Required pages:
- /tasks
- /tasks/new
- /tasks/[id]

Required UI:
- task list
- create task form
- task detail
- status display
- events timeline
- artifacts viewer

### Phase 4: Manual opencode workflow

Do not automate opencode yet.

User will:
1. Create task in dashboard
2. Ask opencode to plan
3. Paste plan into dashboard
4. Approve manually
5. Ask opencode to implement one phase
6. Review diff
7. Commit changes
