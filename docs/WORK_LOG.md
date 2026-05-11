# Work Log

## Historical Log: Completed Phases

### Phase A: Solid Task Control Plane

**Status**: ✅ Complete

**Goal**: Build a control plane for AI-assisted software development.

**Delivered**:
- Tasks CRUD API
- Task events API
- Task artifacts API
- Frontend: /tasks, /tasks/new, /tasks/[id]
- Database: `ai_tasks`, `ai_task_events`, `ai_task_artifacts`

**Migrations**:
- `db/migrations/000001_init.sql`

---

### Phase A.1: Cleanup/Fixes

**Status**: ✅ Complete

**Delivered**:
- Bug fixes and validation
- Test coverage improvements

---

### Phase B.1: Agent Orchestration Database/Backend API Skeleton

**Status**: ✅ Complete

**Goal**: Database and backend API skeleton for agent orchestration.

**Delivered**:
- Agent profiles API (CRUD)
- Agent teams API (CRUD + members)
- Agent runs API (create, list, get)
- Run steps API (create, list, update status)
- Messages API
- Approvals API
- Tool calls API

**Database Tables Added**:
- `agent_profiles`
- `agent_teams`
- `agent_team_members`
- `agent_runs`
- `agent_run_steps`
- `agent_messages`
- `agent_tool_calls`
- `human_approvals`

**Migrations**:
- `db/migrations/000002_agent_orchestration.sql`

---

### Phase B.1.1: Validation and Backend Test Coverage

**Status**: ✅ Complete

**Delivered**:
- Backend handler tests
- Repository tests
- All tests passing

---

### Phase B.2.1: Frontend Agent Profiles/Teams

**Status**: ✅ Complete

**Delivered**:
- Agent profiles page (`/agents`)
- Create profile page (`/agents/profiles/new`)
- Agent teams page
- Create team page (`/agents/teams/new`)
- Team detail page (`/agents/teams/[id]`)

---

### Phase B.2.1.1: Team Member Position Field

**Status**: ✅ Complete

**Delivered**:
- Fixed team member position field handling
- Position field properly used for ordering

---

### Phase B.2.2: Task Agent Runs + Run Steps UI

**Status**: ✅ Complete

**Delivered**:
- Agent run detail page (`/agent-runs/[id]`)
- Run steps list
- Add step form
- Step status selector
- Status badges

---

### Phase B.2.3: Messages, Approvals, Tool Calls UI

**Status**: ✅ Complete

**Delivered**:
- Agent Messages Panel (list + add form)
- Human Approvals Panel (list + status selector)
- Tool Calls Panel (list + status selector)
- Add message form
- Add approval form
- Add tool call form
- 2-column grid layout in run detail

---

### Phase B.2.3.1: Frontend API Tests and Import Cleanup

**Status**: ✅ Complete

**Delivered**:
- Added 8 new API tests for messages, approvals, tool calls
- Frontend tests: 52 passing (was 44)
- MSW mock handlers for all new endpoints
- Cleaned up unused imports in 5 AgentRuns components

---

## Phase C: Auto Orchestration MVP

### Phase C.1.1: Runtime Correctness Fixes

**Status**: ✅ Complete

**Goal**: Fix runtime correctness issues in Phase C orchestration MVP before hardening.

**Issues Fixed**:
1. **Background context cancellation bug** - Orchestrator now uses `context.Background()` instead of request context
2. **Cancel not stopping execution** - Added `running map[string]context.CancelFunc` with `sync.Mutex` for per-run cancellation
3. **AgentRunStepRepository scan bugs** - Fixed missing `step_type` column in `GetByID`, `UpdateOutput`, `MarkStarted`, `MarkCompleted`, `MarkFailed`
4. **Duplicate step creation** - `StartRun()` now checks `GetStepsByRunID()` before creating steps from team
5. **Step type mapping** - Added `mapMemberRoleToStepType()` with proper role mapping (planner→plan, etc.)
6. **Frontend polling useEffect loop** - Split initial data load and polling into separate `useEffect`s; polls only when `run.status === 'running'`

### Phase C.1.2: Orchestration Hardening

**Status**: ✅ Complete

**Goal**: Harden existing Phase C orchestration engine for MVP/demo use.

**Issues Fixed**:
1. **Execute only pending steps** (`orchestrator.go` `executeRun()`):
   - Categorizes steps at start: `pending`, `completed`, `skipped`, `failed`
   - Executes ONLY `status == "pending"` steps
   - Preserves `completed`/`skipped` steps untouched
   - Fails immediately if ANY `failed` steps exist at start
   - If no pending steps (only completed/skipped), marks run `completed` with descriptive summary
   - Includes outputs from already `completed` steps in `previousOutputs` context

2. **Atomic DB status transition for concurrent start safety** (`repository/agent_run.go`):
   - Added `UpdateRunStatusIfIn(id, newStatus, allowedStatuses)` method
   - Uses Postgres `status = ANY($3)` with `pq.Array()` for atomic check+update
   - Returns clear error if no row updated (status not in allowed set)
   - `StartRun()` now registers in running map ONLY AFTER successful DB transition
   - Removed racy in-memory-only check; DB is source of truth

3. **Cancel not overwritten by context.Canceled** (`orchestrator.go` `executeRun()`):
   - All error paths now check `ctx.Err()` FIRST
   - If context is cancelled (`ctx.Err() != nil`), marks run as `cancelled` (NOT `failed`)
   - Writes summary: "Run cancelled."
   - Applies to: LLM errors, `MarkStepStarted` errors, `GetProfileByID` errors, `MarkStepCompleted` errors

4. **Clearer terminal summaries** (`orchestrator.go`):
   - **Completed with executed steps**: "Run completed successfully. Executed X steps."
   - **Completed with no pending steps**: "Run completed: no pending steps to execute. (Completed: X, Skipped: Y, Failed: 0)"
   - **Failed due to existing failed step**: "Run failed: found existing failed step(s). Cannot continue execution."
   - **Failed during execution**: "Run failed: error during step execution." (sanitized, no raw errors)
   - **Cancelled**: "Run cancelled."

5. **Test coverage** (`orchestrator_test.go`):
   - `TestExecuteRun_CompletedStepsNotRerun`: Completed steps preserved, pending executed
   - `TestExecuteRun_SkippedStepsNotRerun`: Skipped steps preserved, pending executed
   - `TestExecuteRun_ExistingFailedStepFailsRun`: Fails immediately with summary
   - `TestExecuteRun_NoPendingStepsCompletes`: Completes without LLM calls, counts summary
   - `TestExecuteRun_CancelledDuringLLM_MarksCancelled`: Context cancel → `cancelled` status, not `failed`
   - `TestStartRun_AtomicStatusTransition`: Fails when status not allowed; running map only populated on success

---

## Current Status

### Next Step: User Demo / Further Phases

**Phase C**: Auto Orchestration MVP - Core Complete

**Completed Phases in C**:
- **C.1.1**: Runtime Correctness Fixes
- **C.1.2**: Orchestration Hardening

**Core Functionality Now Available**:
- ✅ Start/cancel endpoints (`POST /agent-runs/:id/start`, `POST /agent-runs/:id/cancel`)
- ✅ Atomic status transitions (prevents concurrent double-starts)
- ✅ Only pending steps executed (completed/skipped preserved)
- ✅ Existing failed steps fail the run immediately
- ✅ Context cancellation properly marks as `cancelled` (not `failed`)
- ✅ Clear terminal summaries
- ✅ Dashboard polling (every 3s while `running`)
- ✅ Start/Cancel buttons in UI

**Limitations (Intentional for MVP)**:
- No WebSocket (polling only)
- No distributed queue (in-memory goroutines only)
- No real dangerous tool execution
- No automatic codebase modification
- No authentication

---

## Open Risks (Resolved)

1. **AGENTS.md previously blocked AI automation**
   - OLD: "Do not add AI automation yet."
   - FIXED: Updated to allow AI automation in Phase C with proper constraints
   - New rule: "AI automation is allowed only in Phase C work, must be scoped, configurable, tested, and must not hardcode secrets."

2. **No shared memory files existed**
   - FIXED: Created docs files for project memory:
     - `docs/PROJECT_BRIEF.md`
     - `docs/CURRENT_STATUS.md`
     - `docs/PHASE_C_PLAN.md`
     - `docs/AGENT_TEAM.md`
     - `docs/WORK_LOG.md`
   - Created `.opencode/agents/*.md` for OpenCode agent team

3. **Multi-agent work needs strict ownership**
   - FIXED: `docs/AGENT_TEAM.md` defines clear ownership boundaries
   - @leader coordinates
   - Each agent stays within their scope

4. **Phase C runtime correctness issues**
   - FIXED: Phase C.1.1 addressed all 7 known issues

5. **Phase C hardening gaps**
   - FIXED: Phase C.1.2 addressed all 5 hardening issues with test coverage

---

## Latest Verification (After Phase C.1.2)

### Backend
- `go test ./...` → All passing
- Orchestrator tests: ~40 tests covering all hardening scenarios

### Frontend
- `npm run test:run` → 54 tests passing
- `npm run lint` → No warnings/errors
- `npm run build` → Success (9 static pages generated)

### Database
- 2 migrations applied:
  - `000001_init.sql`
  - `000002_agent_orchestration.sql`

### Key Endpoints Available
See `docs/CURRENT_STATUS.md` for full list.

### Key Phase C Endpoints
- `POST /agent-runs/:id/start` - Start run execution
- `POST /agent-runs/:id/cancel` - Cancel run execution

---

## Quick Reference

| Document | Purpose |
|----------|---------|
| `docs/PROJECT_BRIEF.md` | Product vision, goals, non-goals, success criteria |
| `docs/CURRENT_STATUS.md` | What exists, what doesn't, API reference |
| `docs/PHASE_C_PLAN.md` | Detailed plan for auto orchestration |
| `docs/AGENT_TEAM.md` | Agent definitions, ownership, coordination |
| `docs/WORK_LOG.md` | This file - history and current status |
| `AGENTS.md` | Project-level rules and phase definition |
| `opencode.json` | OpenCode configuration |
| `.opencode/agents/*.md` | OpenCode agent team definitions |

---

## Git Commands Reference

```bash
# Check status
git status

# View diff
git diff

# Add all changes
git add .

# Commit
git commit -m "message"
```
