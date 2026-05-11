# Current Status

## What Exists Now (as of Phase C.1.3)

### Backend Stack

- **Language**: Go 1.22+
- **Router**: gorilla/mux
- **Database**: Postgres via database/sql + lib/pq
- **Pattern**: Handler → Repository → Model
- **Configuration**: JSON config + environment variables

### Frontend Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **UI Components**: shadcn/ui + Tailwind CSS
- **Testing**: Vitest + MSW (Mock Service Worker)
- **API Client**: Custom fetch wrapper with `{ data: ... }` / `{ error: ... }` response format

### Database Migrations

| File | Purpose |
|------|---------|
| `db/migrations/000001_init.sql` | Tasks, Events, Artifacts (Phase 1) |
| `db/migrations/000002_agent_orchestration.sql` | Agent orchestration tables (Phase B.1) |

### Key Database Tables

```
ai_tasks              - Development tasks
ai_task_events        - Task events/notes
ai_task_artifacts     - Task artifacts/code

agent_profiles        - Agent definitions (name, role, system_prompt, default_model)
agent_teams           - Team definitions
agent_team_members    - Team membership (team_id, profile_id, member_role, position)
agent_runs            - Execution runs (task_id, team_id, status, goal, summary)
agent_run_steps       - Run steps (run_id, profile_id, step_type, status, title, instructions, output)
agent_messages        - Messages during run (run_id, step_id, profile_id, role, content, metadata)
agent_tool_calls      - Tool execution records (run_id, step_id, tool_name, input, output, status)
human_approvals       - Approval gates (task_id, run_id, step_id, approval_type, status)
```

### Key Backend Endpoints

#### Tasks
- `GET /health` - Health check
- `POST /tasks` - Create task
- `GET /tasks` - List tasks
- `GET /tasks/:id` - Get task detail
- `PATCH /tasks/:id/status` - Update task status
- `PATCH /tasks/:id/plan` - Update task plan
- `PATCH /tasks/:id/review-notes` - Update review notes

#### Events & Artifacts
- `POST /tasks/:id/events` - Create event
- `GET /tasks/:id/events` - List events
- `POST /tasks/:id/artifacts` - Create artifact
- `GET /tasks/:id/artifacts` - List artifacts

#### Agent Profiles
- `POST /agent-profiles` - Create profile
- `GET /agent-profiles` - List profiles
- `GET /agent-profiles/:id` - Get profile detail

#### Agent Teams
- `POST /agent-teams` - Create team
- `GET /agent-teams` - List teams
- `GET /agent-teams/:id` - Get team detail
- `POST /agent-teams/:id/members` - Add team member
- `GET /agent-teams/:id/members` - List team members

#### Agent Runs
- `POST /tasks/:id/agent-runs` - Create run
- `GET /tasks/:id/agent-runs` - List runs by task
- `GET /agent-runs/:id` - Get run detail
- `POST /agent-runs/:id/start` - Start run execution (Phase C)
- `POST /agent-runs/:id/cancel` - Cancel run execution (Phase C)

#### Agent Run Steps
- `POST /agent-runs/:id/steps` - Create step
- `GET /agent-runs/:id/steps` - List steps
- `PATCH /agent-run-steps/:id/status` - Update step status

#### Agent Messages
- `POST /agent-runs/:id/messages` - Create message
- `GET /agent-runs/:id/messages` - List messages

#### Human Approvals
- `POST /agent-runs/:id/approvals` - Create approval
- `GET /agent-runs/:id/approvals` - List approvals
- `PATCH /human-approvals/:id/status` - Update approval status

#### Agent Tool Calls
- `POST /agent-runs/:id/tool-calls` - Create tool call
- `GET /agent-runs/:id/tool-calls` - List tool calls
- `PATCH /agent-tool-calls/:id/status` - Update tool call status

### Key Frontend Pages

| Route | Page |
|-------|------|
| `/` | Home (placeholder) |
| `/tasks` | Task list |
| `/tasks/new` | Create task form |
| `/tasks/[id]` | Task detail |
| `/agents` | Agent profiles list |
| `/agents/profiles/new` | Create agent profile |
| `/agents/teams/new` | Create agent team |
| `/agents/teams/[id]` | Team detail (with members) |
| `/agent-runs/[id]` | Run detail (steps, messages, approvals, tool calls) |

### Key Frontend Components

```
components/
  AgentRuns/
    AgentRunDetail.tsx         - Run detail page with panels
    AgentMessagesPanel.tsx     - Messages list + add form
    HumanApprovalsPanel.tsx    - Approvals list + status selector
    ToolCallsPanel.tsx         - Tool calls list + status selector
    AgentRunBadges.tsx         - Status badges
    AddAgentMessageForm.tsx    - Add message form
    AddHumanApprovalForm.tsx   - Add approval form
    AddToolCallForm.tsx        - Add tool call form
    ApprovalStatusSelector.tsx - Approval status dropdown
    ToolCallStatusSelector.tsx - Tool call status dropdown
```

### Phase C Capabilities (Now Implemented)

#### Orchestrator Core (services/api/internal/service/orchestrator.go)
- **StartRun**: Atomically transitions run status, creates steps from team if needed, registers in running map
- **executeRun**: Executes only `pending` steps; preserves `completed`/`skipped` steps; fails immediately on existing `failed` steps
- **CancelRun**: Calls stored cancel function, updates status to `cancelled`
- **State safety**: Atomic database status transition using `status = ANY($3)`; concurrent starts blocked at DB level
- **Context cancellation handling**: All error paths check `ctx.Err()` first; marks as `cancelled` (not `failed`) when context is done
- **Step type mapping**: `planner`→`plan`, `implementer/backend/frontend/database`→`implement`, `reviewer`→`review`, `qa`→`test`, default→`implement`
- **Clear summaries**: Terminal outcomes have descriptive summaries in `run.summary`

#### Task Context Integration (Phase C.1.3)
- **Task context included in LLM prompts**:
  - Task title, description, plan, review_notes
  - Run goal
  - Step instructions
  - Previous step outputs
- **Structured format**: `=== TASK CONTEXT ===`, `=== RUN GOAL ===`, `=== STEP INSTRUCTIONS ===`, `=== PREVIOUS STEP OUTPUTS ===`
- **Graceful handling**:
  - If `run.TaskID` is empty: continues without task context
  - If task load fails: marks run `failed` with sanitized summary ("Run failed: failed to load task context.")

#### LLM Provider Interface (services/api/internal/llm/)
- `LLMProvider` interface with `ChatCompletion(ctx, messages)`
- **FakeProvider**: No API key required, returns deterministic mock responses for demos
- **OpenAIProvider**: Configurable via environment variables
- **Environment configuration**:
  - `LLM_PROVIDER`: `openai` or `fake` (default: `openai`)
  - `LLM_API_KEY`: Required when `LLM_PROVIDER=openai`
  - `LLM_MODEL`: Model name (default: `gpt-4`)
  - `LLM_BASE_URL`: Custom API base URL for OpenAI-compatible endpoints
  - `LLM_TIMEOUT`: Request timeout (Go duration format, default: `60s`)

#### Dashboard Controls (apps/web/)
- **Start Run** / **Cancel Run** buttons in run detail page
- **Polling**: Auto-refreshes run data every 3 seconds while run status is `running`
- **Refresh Button**: Manual Refresh button in Agent Messages panel
- **API functions**: `startAgentRun()`, `cancelAgentRun()`

### Latest Verification Status

Last verified after Phase C.1.3:

- **Backend tests**: `go test ./...` → All passing
- **Backend test count**: ~43 orchestrator tests (role mapping, status validation, pending-only execution, existing-failed fail-fast, cancellation handling, atomic transitions, multi-run concurrency, task context integration)
- **Frontend tests**: `npm run test:run` → 54 tests passing
- **Frontend lint**: `npm run lint` → No ESLint warnings or errors
- **Frontend build**: `npm run build` → Success (9 static pages generated)
- **Config validation**: `jq empty opencode.json` → Valid JSON

---

## What Does NOT Exist Yet

### Not Implemented (Out of Scope for MVP)

- `POST /agent-runs/:id/pause` - Pause execution
- `POST /agent-runs/:id/resume` - Resume execution

### Limitations (By Design for MVP)

#### No WebSocket Real-time Updates
- Dashboard uses polling (3-second interval) only while run is `running`
- No WebSocket or SSE push notifications

#### No Distributed Queue
- Execution is in-memory goroutine per run
- No Redis, RabbitMQ, or external job queue
- Runs don't survive process restart

#### No Real Dangerous Tool Execution
- Tool calls are database records only
- No actual execution of:
  - File write (read may be added later)
  - Bash commands
  - Git operations that modify state
  - Web requests that modify external systems

#### No Automatic Codebase Modification by AI
- All AI execution must be explicitly planned, scoped, and approved
- Phase C MVP focuses on orchestration, not actual code modification

#### No Authentication
- No user accounts, login, OAuth, JWT, or permission systems

#### No LLM API Keys in Repository
- All config must come from environment variables
- `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL`, `LLM_BASE_URL`

---

## Data Models Reference

### Status Enums Used

**TaskStatus**: `pending`, `planning`, `approved`, `in_progress`, `reviewing`, `completed`, `failed`

**AgentRunStatus**: `draft`, `planned`, `waiting_approval`, `approved`, `running`, `paused`, `completed`, `failed`, `cancelled`

**AgentRunStepStatus**: `pending`, `waiting_approval`, `running`, `completed`, `failed`, `skipped`, `cancelled`

**AgentMessageRole**: `system`, `user`, `assistant`, `tool`, `reviewer`

**HumanApprovalStatus**: `pending`, `approved`, `rejected`, `cancelled`

**AgentToolCallStatus**: `recorded`, `approved`, `rejected`, `completed`, `failed`

### Step Types

Step types used in UI and tests:
- `plan` - Planning phase
- `implement` - Implementation phase
- `review` - Review phase
- `test` - Testing phase
