# Current Status

## What Exists Now (as of Phase C.3)

### Backend Stack

- **Language**: Go 1.22+
- **Router**: gorilla/mux
- **Database**: Postgres via database/sql + lib/pq
- **Pattern**: Handler → Repository → Model
- **Configuration**: JSON config + environment variables
- **Tool Config**: `WORKSPACE_ROOT`, `TOOL_READ_MAX_BYTES` (default 1MB), `TOOL_SEARCH_MAX_RESULTS` (default 50), `TOOL_MAX_ITERATIONS` (default 10), `TOOL_WRITE_MAX_BYTES` (default 1MB), `TOOL_BASH_TIMEOUT` (default 30s), `TOOL_BASH_BLOCKED_COMMANDS`, `TOOL_REQUIRE_APPROVAL`, `TOOL_CHOICE`

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
- `POST /agent-runs/:id/resume` - Resume paused run (Phase C.4)

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
    HumanApprovalsPanel.tsx    - Approvals list + approve/reject buttons + tool input/output display
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
- `LLMProvider` interface with:
  - `ChatCompletion(ctx, messages)` — backward compatible text-only completion
  - `ChatCompletionWithTools(ctx, messages, tools)` — native function calling with tool definitions
- `ChatCompletionResponse` — returns both `Content` (text) and `ToolCalls` (native structured)
- **FakeProvider**: No API key required, returns deterministic mock responses; `ChatCompletionWithTools` delegates to text (backward compat)
- **OpenAIProvider**: `ChatCompletionWithTools` sends `tools` parameter via OpenAI API; parses native `tool_calls` from response; `ChatCompletion` delegates to `ChatCompletionWithTools` with no tools (no code duplication)
- **Environment configuration**:
  - `LLM_PROVIDER`: `openai` or `fake` (default: `openai`)
  - `LLM_API_KEY`: Required when `LLM_PROVIDER=openai`
  - `LLM_MODEL`: Model name (default: `gpt-4`)
  - `LLM_BASE_URL`: Custom API base URL for OpenAI-compatible endpoints
  - `LLM_TIMEOUT`: Request timeout (Go duration format, default: `60s`)

#### Dashboard Controls (apps/web/)
- **Start Run** / **Cancel Run** / **Resume Run** buttons in run detail page
- **Polling**: Auto-refreshes run data every 3 seconds while run status is `running` or `paused`
- **Refresh Button**: Manual Refresh button in Agent Messages panel
- **API functions**: `startAgentRun()`, `cancelAgentRun()`, `resumeAgentRun()`

#### Read-Only Tool Execution (Phase C.2.1 - `services/api/internal/tool/`)

**New tool package** with 3 read-only tools:

| Tool | Input | Output | Safety |
|------|-------|--------|--------|
| `list_files` | `{path, depth}` | File/dir listing | Blocked: `.git`, `node_modules`, `.env`, etc. Depth-controlled |
| `read_file` | `{path}` | Text content (truncated) | Binary detection, extension check, max bytes limit |
| `search_code` | `{pattern, path?, is_regex?}` | Matching lines + snippets | Regex support, max results limit, skips blocked paths |

**Path safety rules**:
- Rejects absolute paths and `../` traversal
- Validates path stays within `WORKSPACE_ROOT`
- Blocks: `.env`, `.env.*`, `.git`, `node_modules`, `.next`, `dist`, `build`, `vendor`
- Binary detection by extension + null byte check
- Symlink escape detection via `filepath.EvalSymlinks`

**Multi-Turn Tool Feedback Loop (Phase C.2.2)**:
- LLM can now iterate: response → tool_calls → execute → feed results back → re-query LLM → repeat
- Loop terminates when LLM response has no `tool_calls` (final answer) or `MaxToolIterations` is reached
- All tool results are stored as `agent_messages` with `role: "tool"` AND appended to LLM conversation context
- Step output = final non-tool-call LLM response (or fallback `"Step completed."` if max iterations reached)
- Configurable via `TOOL_MAX_ITERATIONS` env var (default: 10)

**Integration**:
- LLM can trigger tools via JSON format: `{"tool_calls":[{"tool_name":"read_file","input":{"path":"..."}}]}`
- Tool calls recorded in `agent_tool_calls` table (status: `completed`/`failed`)
- Tool results stored as `agent_messages` with `role: "tool"` AND fed back into LLM context
- Multi-turn loop: each LLM call includes full conversation history (system + user + assistant + tool messages)
- Tool errors do not crash the server

**Tool Approval Gates (Phase C.2.3 + C.4)**:
- Configurable via `TOOL_REQUIRE_APPROVAL` env var (comma-separated list of tool names, e.g., `write_file,bash`)
- Before executing each tool call, the orchestrator checks if the tool requires approval
- If approval required:
  - Creates `human_approval` record with `status: pending`, `approval_type: "tool:<name>"`
  - Creates `agent_tool_call` with `status: "recorded"` (not executed)
  - **Pauses the run** (status: `paused`, step: `waiting_approval`)
  - Exits the goroutine — run is blocked until resumed
- On resume: auto-resume check finds approved tool call, executes it, feeds result back to LLM
- If no approval required: tool executes normally (existing behavior)
- Frontend approvals panel polls every 3s during running or paused state
- Approve/Reject buttons auto-resume the run

**Native LLM Function Calling (Phase C.2.4)**:
- Orchestrator now calls `ChatCompletionWithTools()` with registered tool definitions (`list_files`, `read_file`, `search_code`)
- **OpenAIProvider** uses native `tools` API parameter: sends `ToolDefinition` as OpenAI `function` objects, parses `tool_calls` from response
- **FakeProvider** falls back to text-based tool call parsing (backward compatible)
- Orchestrator handles two paths:
  - **Native**: Uses `resp.ToolCalls` directly when LLM returns structured tool calls
  - **Fallback**: Parses text response with `tool.ParseToolCalls()` for backward compatibility
- `ToolCall` converted to `tool.ToolCallRequest` before execution, then to `tool.ToolCallResponse` after
- `ChatCompletion` method remains for backward compatibility (delegates to `ChatCompletionWithTools` with no tools)
- `tool_choice` parameter configurable via `TOOL_CHOICE` env var (`"none"`, `"auto"`, `"required"`; default: `"auto"`)

**New config env vars**:
- `WORKSPACE_ROOT` - Project root (defaults to current working directory)
- `TOOL_READ_MAX_BYTES` - Max bytes to read per file (default: 1048576)
- `TOOL_SEARCH_MAX_RESULTS` - Max search results (default: 50)
- `TOOL_MAX_ITERATIONS` - Max tool call iterations per step (default: 10)
- `TOOL_REQUIRE_APPROVAL` - Comma-separated tool names requiring approval before execution (default: empty)
- `TOOL_CHOICE` - Tool calling mode: `none`, `auto`, `required` (default: `auto`)
- `TOOL_WRITE_MAX_BYTES` - Max bytes for write/edit files (default: 1048576)
- `TOOL_BASH_TIMEOUT` - Timeout in seconds for bash commands (default: 30)
- `TOOL_BASH_BLOCKED_COMMANDS` - Additional comma-separated blocked bash patterns

### Added Phase C.3 Capabilities

#### Write Tool Execution (Phase C.3)

The tool package now has 4 write/execution tools:

| Tool | Input | Output | Safety |
|------|-------|--------|--------|
| `write_file` | `{path, content}` | Path, size, type | Binary detection, null byte check, max size, parent dir auto-create, blocked path reject |
| `edit_file` | `{path, old_string, new_string}` | Path, replaced_count, type | Find-and-replace, binary reject, null byte check, no-match error, blocked path reject |
| `bash` | `{command, timeout?}` | Stdout, stderr, exit_code | 40 blocked patterns (sudo, dd, mkfs, chown, pipe-to-sh, etc.), timeout, 100KB output limit |
| `git` | `{args: ["status", "--short"]}` | Stdout, stderr, exit_code | 18 blocked commands (push, reset, clean, rebase, merge, gc, etc.), config write blocks, 60s timeout |

**New files in tool package**:
- `write_file.go` + `write_file_test.go` (9 tests)
- `edit_file.go` + `edit_file_test.go` (9 tests)
- `bash.go` + `bash_test.go` (8 tests)
- `git.go` + `git_test.go` (10 tests)

**Safety enhancements** (`safety.go`):
- `ValidateWritePath()` — like `ValidatePath` but allows non-existent files, with anti-symlink-escape walk
- `IsBashCommandAllowed()` — 40 blocked patterns pipeline
- `IsGitCommandAllowed()` — 18 blocked git commands + subcommand flag analysis

### Latest Verification Status

Last verified after Phase C.4 (Pause/Resume):

- **Backend tests**: `go test ./...` → All passing (128+ test cases)
  - `internal/tool`: 83 tests
  - `internal/service`: All passing (15+ tests including pause/resume)
  - `internal/handler`: All passing (10+ tests including resume handler)
  - `internal/config`: All passing
  - `internal/server`: All passing
- **Frontend tests**: `npm run test:run` → 54 tests passing
- **Frontend lint**: `npm run lint` → No ESLint warnings or errors
- **Frontend build**: `npm run build` → Success (9 static pages generated)
- **Config validation**: `jq empty opencode.json` → Valid JSON

---

## What Does NOT Exist Yet

### Not Implemented (Out of Scope for MVP)

- `POST /agent-runs/:id/pause` - Pause execution (handled automatically by orchestrator on approval gate)

### Phase C.3 (Completed)

- **Write tool execution** (`write_file`, `edit_file`, `bash`, `git`) — ✅ Implemented with full safety guards
- **Approval Auto-Resume** — ✅ When a human approves a tool via PATCH, the orchestrator automatically executes the tool and feeds the result into the LLM conversation, without manual re-run
- **Approvals Dashboard** — ✅ Rich UI with Approve/Reject buttons, tool input as formatted JSON, tool call result correlation, inline error handling
- **Function-specific `tool_choice`** — only string values supported (`none`/`auto`/`required`); cannot force a specific tool yet

### Limitations (By Design for MVP)

#### No WebSocket Real-time Updates
- Dashboard uses polling (3-second interval) only while run is `running`
- No WebSocket or SSE push notifications

#### No Distributed Queue
- Execution is in-memory goroutine per run
- No Redis, RabbitMQ, or external job queue
- Runs don't survive process restart

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
