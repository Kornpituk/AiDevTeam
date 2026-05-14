# AiDevT

**Agent Orchestration Engine + Dashboard for AI-assisted software development.**

## Goal

Build an end-to-end system that allows:
1. Users to define reusable agent profiles
2. Users to define agent teams
3. Users to start an agent run for a task
4. Backend to automatically execute steps using an orchestration engine
5. Dashboard to track run status, steps, messages, approvals, tool calls, errors, and summaries

## Stack

- Frontend: Next.js + TypeScript in `apps/web`
- Backend: Go in `services/api`
- Database: Postgres
- Migrations: SQL files in `db/migrations`

## Project Structure

- `apps/web`: Next.js frontend
- `services/api`: Go backend API
- `db/migrations`: SQL migrations
- `docs`: Project documentation
- `.opencode/agents`: OpenCode agent team definitions

## Development Rules

- Always read docs/PROJECT_BRIEF.md, docs/CURRENT_STATUS.md, and docs/AGENT_TEAM.md first.
- Always plan before editing files.
- Keep patches small.
- Do not implement multiple phases at once.
- Do not edit `.env` files.
- Do not add auth yet unless explicitly requested.
- Do not create destructive database migrations.
- Do not modify old migrations after they are applied.
- Show git diff after changes.
- Backend changes should run `go test ./...`.
- Frontend changes should run `npm run lint` and `npm run build` when available.

## Phase C: AI Automation Rule

**AI automation is allowed only in Phase C work, must be scoped, configurable, tested, and must not hardcode secrets.**

- No hardcoded API keys or tokens anywhere
- All LLM config must come from environment variables
- Execution must be testable without real API calls
- Guardrails must exist: timeouts, approval gates, error recovery
- Show git diff after all automation-related changes

## Multi-Agent Work Rules

1. **Read docs first**: Always read `docs/PROJECT_BRIEF.md` and `docs/CURRENT_STATUS.md` before work.
2. **One leader controls scope**: @leader decides what to work on and when to switch phases.
3. **Stay within ownership**: Each agent must stay within their defined boundaries (see `docs/AGENT_TEAM.md`).
4. **Avoid overlapping edits**: Coordinate if multiple agents need to touch the same file.
5. **No old migration changes**: Never modify existing `db/migrations/*.sql` files.
6. **No auth unless requested**: Do not add authentication unless explicitly requested.
7. **No destructive migrations**: Do not drop tables or delete data irreversibly.
8. **No hardcoded secrets**: Never commit API keys or credentials.
9. **Show diff after changes**: Always show git diff after making changes.
10. **Run relevant tests**: Backend: `go test ./...`, Frontend: `npm run test:run`, `npm run lint`, `npm run build`.

## Completed Phases (Historical)

### Phase A: Solid Task Control Plane

Database and API for tasks, events, artifacts.
- Tasks CRUD
- Task events
- Task artifacts

### Phase A.1: Cleanup/Fixes

Bug fixes and validation improvements.

### Phase B.1: Agent Orchestration Database/Backend API Skeleton

Database schema and backend API skeleton for agent orchestration:
- Agent profiles API
- Agent teams + members API
- Agent runs API
- Run steps API
- Messages API
- Approvals API
- Tool calls API

**Tables added**: `agent_profiles`, `agent_teams`, `agent_team_members`, `agent_runs`, `agent_run_steps`, `agent_messages`, `agent_tool_calls`, `human_approvals`

### Phase B.1.1: Validation and Backend Test Coverage

Backend handler tests, repository tests.

### Phase B.2.1: Frontend Agent Profiles/Teams

Frontend pages:
- `/agents` - Agent profiles list
- `/agents/profiles/new` - Create profile
- `/agents/teams/new` - Create team
- `/agents/teams/[id]` - Team detail

### Phase B.2.1.1: Team Member Position Field

Fixed team member position field handling.

### Phase B.2.2: Task Agent Runs + Run Steps UI

Frontend:
- `/agent-runs/[id]` - Run detail page
- Run steps list
- Add step form
- Step status selector

### Phase B.2.3: Messages, Approvals, Tool Calls UI

Frontend panels in run detail:
- Agent messages panel (list + add form)
- Human approvals panel (list + status selector)
- Tool calls panel (list + status selector)
- 2-column grid layout

### Phase B.2.3.1: Frontend API Tests and Import Cleanup

- Added 8 new API tests for messages, approvals, tool calls
- Cleaned up unused imports in AgentRuns components
- Frontend: 52 tests passing, lint passing, build passing

## Completed Phases in C

### Phase C: Auto Orchestration MVP ✅ COMPLETE

Goal: Enable automatic execution of agent runs.

**All sub-phases completed**:
- **C.1**: Orchestrator Core ✅ — state transitions, step creation from team
- **C.2**: LLM Provider Interface ✅ — abstraction for LLMs (OpenAI + Fake)
- **C.3**: Execution Loop ✅ — step execution, multi-turn tool loop, native function calling
- **C.4**: Dashboard Controls ✅ — start/cancel/resume UI, polling, pause/approve
- **C.5**: Guardrails and Tests ✅ — +27 backend tests, +79 frontend tests, stress test, edge cases

**Key endpoints added**:
- `POST /agent-runs/:id/start` — Start run execution
- `POST /agent-runs/:id/cancel` — Cancel run execution
- `POST /agent-runs/:id/resume` — Resume paused run

**Total test coverage**:
- Backend: 116 tests passing across 6 packages
- Frontend: 133 tests passing, lint clean, build successful

### Phase D.1: WebSocket Real-Time Updates ✅ COMPLETE

Goal: Replace polling with WebSocket push for live updates, with polling fallback.

**Delivered**:
- `internal/ws/` package — Hub, Client, Handler with room-based broadcast
- `GET /ws/agent-runs/{id}` — WebSocket upgrade endpoint
- 16 broadcast points in orchestrator (run, step, message, approval, tool_call updates)
- Frontend `useWebSocket` hook with auto-reconnect (exponential backoff)
- Polling kept as fallback when WebSocket disconnects
- Green "Live" badge indicator
- Frontend tests: 143 (+10 WebSocket tests), lint clean, build successful

## What's Next

Phase C and D.1 are fully complete. Future work could include:
- Distributed job queue
- Authentication/authorization
- Additional LLM providers (Anthropic, Gemini)
- Advanced tool implementations

## Phase 4: Manual Workflow (Legacy/Fallback)

This was the original manual approach, kept as historical context:

User will:
1. Create task in dashboard
2. Ask opencode to plan
3. Paste plan into dashboard
4. Approve manually
5. Ask opencode to implement one phase
6. Review diff
7. Commit changes

## Project Docs

For full documentation, see:
- `docs/PROJECT_BRIEF.md` - Product vision, goals, non-goals
- `docs/CURRENT_STATUS.md` - What exists, what doesn't, API reference
- `docs/PHASE_C_PLAN.md` - Detailed plan for auto orchestration
- `docs/AGENT_TEAM.md` - Agent definitions, ownership, coordination
- `docs/WORK_LOG.md` - History, current status, next steps
