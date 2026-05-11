# Current Status

## What Exists Now (as of Phase B.2.3.1)

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

### Latest Verification Status

Last verified after Phase B.2.3.1:

- **Backend tests**: `go test ./...` → All passing (cached)
- **Frontend tests**: `npm run test:run` → 52 tests passing
  - `components/ui/badge.test.ts` (8 tests)
  - `lib/utils.test.ts` (5 tests)
  - `components/ui/card.test.tsx` (4 tests)
  - `components/ui/button.test.tsx` (7 tests)
  - `lib/api.test.ts` (28 tests) - covers messages, approvals, tool calls
- **Frontend lint**: `npm run lint` → No ESLint warnings or errors
- **Frontend build**: `npm run build` → Success (9 static pages generated)

---

## What Does NOT Exist Yet (Phase C)

### Auto Orchestration Engine

1. **No orchestrator core**: No component that:
   - Takes a run and determines what steps to create
   - Reads team members and creates steps from them
   - Manages step transitions based on dependencies

2. **No LLM provider interface**: No abstraction for calling LLMs (OpenAI, Anthropic, etc.)

3. **No execution loop**: No background worker/poller that:
   - Checks for pending/running runs
   - Executes steps in order
   - Calls LLM for step execution
   - Records messages and tool calls
   - Updates step/run status

### Missing Endpoints

- `POST /agent-runs/:id/start` - Start execution of a run
- `POST /agent-runs/:id/pause` - Pause execution
- `POST /agent-runs/:id/resume` - Resume execution
- `POST /agent-runs/:id/cancel` - Cancel execution

### Missing Dashboard Controls

- No "Start Run" button in UI
- No "Pause" / "Resume" / "Cancel" controls
- No execution progress visualization beyond status badges
- No polling for live updates (optional for MVP)

### Missing Configuration

- No LLM provider config in backend
- No environment variables for API keys (should not be committed)
- No model selection per run/step

### Missing Real Tool Execution

- Tool calls are currently just database records
- No actual execution of:
  - File read/write
  - Bash commands
  - Git operations
  - Web requests

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
