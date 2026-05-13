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

### Phase C.1.2.1: Concurrent Start Ordering Fix

**Status**: ✅ Complete

**Goal**: Fix remaining concurrent start race condition in `StartRun`.

**Problem**:
Previous order in `StartRun`:
1. `GetRunByID`
2. `GetStepsByRunID` + `createStepsFromTeam` (step creation)
3. `UpdateRunStatusIfIn` (atomic transition)

**Race condition**: Two concurrent `POST /agent-runs/:id/start` requests could both:
- See no steps
- Create duplicate steps
- Only one wins atomic transition
- Result: duplicate steps in database

**Fix Applied**:
New safe ordering:
1. `GetRunByID` → get `TeamID`/`Goal`
2. **`UpdateRunStatusIfIn`** → atomic transition to `running` FIRST (DB is source of truth)
3. Register cancel in running map
4. THEN check existing steps & create from team if needed
5. **Cleanup on failure after transition**:
   - If `GetStepsByRunID` fails: `cancel()`, remove from running map, set status `"failed"`, set summary
   - If `createStepsFromTeam` fails: same cleanup
   - No `executeRun` goroutine started on failure
6. Only on full success: `go o.executeRun(...)`

**Tests Added** (`orchestrator_test.go`):
1. `TestStartRun_NoStepsCreatedOnFailedTransition`: Atomic transition fails → 0 steps created, running map empty
2. `TestStartRun_CleansUpOnFailedStepCreation`: Step creation fails after transition → status `"failed"`, summary set, running map empty
3. `TestStartRun_SimulatedConcurrentStart`: Two sequential calls → only 1 set of steps, only 1 "running" transition
4. `TestStartRun_HappyPathNoTeam`: No-team happy path → running map populated, status="running"

---

## Current Status

### Next Step: Further Phases

**Phase C**: Auto Orchestration MVP - Core Complete (including Pause/Resume)

**Completed Phases in C**:
- **C.1.1**: Runtime Correctness Fixes
- **C.1.2**: Orchestration Hardening
- **C.1.2.1**: Concurrent Start Ordering Fix
- **C.1.3**: Demo Readiness + Task Context
- **C.2.1**: Read-Only Tool Execution MVP
- **C.2.2**: Multi-Turn Tool Feedback Loop
- **C.2.3**: Tool Approval Gates
- **C.2.4**: Native LLM Function Calling
- **C.2.5**: `tool_choice` Parameter
- **C.3**: Write Tool Execution (write_file, edit_file, bash, git) ✅
- **C.3.1**: Approval Auto-Resume ✅
- **C.3.2**: Approvals Dashboard ✅
- **C.4**: Pause/Resume ✅ (NEW)

**Core Functionality Now Available**:
- ✅ Start/cancel/resume endpoints
- ✅ Atomic status transitions
- ✅ Only pending steps executed
- ✅ Existing failed steps fail the run immediately
- ✅ Context cancellation properly marks as `cancelled`
- ✅ Clear terminal summaries
- ✅ Dashboard polling (every 3s while `running` or `paused`)
- ✅ Start/Cancel/Resume buttons in UI
- ✅ Read-only tool execution: `list_files`, `read_file`, `search_code`
- ✅ Write tool execution: `write_file`, `edit_file`, `bash`, `git`
- ✅ Tool calls recorded in `agent_tool_calls` table
- ✅ Multi-turn tool feedback loop (iterative LLM + tool execution)
- ✅ Tool approval gates (configurable pre-approval checks per tool)
- ✅ Pause/Resume — approval gate pauses execution; resume auto-executes approved tools
- ✅ Native LLM function calling (OpenAI `tools` API, with text fallback)
- ✅ `tool_choice` parameter (`none`/`auto`/`required` via `TOOL_CHOICE` env var)
- ✅ Bash command safety (40 blocked patterns, timeout, output limits)
- ✅ Git command safety (18 blocked destructive commands)
- ✅ Write path safety (binary reject, null byte check, size limit, blocked dirs)
- ✅ Approval auto-resume (approved tools auto-execute and feed results to LLM)
- ✅ Approvals Dashboard (approve/reject buttons, tool input/output display, auto-refresh, auto-resume)

**Limitations (Intentional for MVP)**:
- No WebSocket (polling only)
- No distributed queue (in-memory goroutines only)
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

6. **Concurrent start ordering race condition**
   - OLD: Step creation happened BEFORE atomic status transition
   - RACE: Two concurrent requests could both create duplicate steps
   - FIXED: Phase C.1.2.1 reordered so atomic transition happens FIRST
   - Added cleanup path if step creation fails after transition

---

### Phase C.1.3: Demo Readiness + Task Context

**Status**: ✅ Complete

**Goal**: Make Phase C Auto Orchestration MVP ready for real demo/use.

**Issues Fixed**:

1. **Task context integrated into orchestration prompts** (`orchestrator.go`):
   - Added `GetTaskByID` to `OrchestratorRepository` interface
   - Wired `TaskRepository` into `OrchestratorRepoImpl`
   - Updated `executeRun` to load task using `run.TaskID`
   - Task context included in LLM user messages with structured format:
     ```
     === TASK CONTEXT ===
     Title: {task.Title}
     Description: {task.Description}
     Plan: {task.Plan}
     Review Notes: {task.ReviewNotes}
     
     === RUN GOAL ===
     {run.Goal}
     
     === STEP INSTRUCTIONS ===
     {step.Instructions}
     
     === PREVIOUS STEP OUTPUTS ===
     {previousOutputs}
     ```
   - If task loading fails: mark run failed with sanitized summary
   - If `run.TaskID` is empty: continue gracefully without task context

2. **LLM config docs updated** (`.env.example`):
   - Documented all LLM env vars: `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL`, `LLM_BASE_URL`, `LLM_TIMEOUT`
   - Clarified `LLM_PROVIDER=fake` for demo without API key

3. **Frontend refresh improvement** (`AgentMessagesPanel.tsx`):
   - Added "Refresh" button to Messages panel
   - Added `refreshKey` state to trigger re-fetch
   - Users can manually refresh messages after run completes

4. **Demo runbook created** (`docs/DEMO_RUNBOOK.md`):
   - Full setup guide: Postgres, migrations, backend, frontend
   - Fake provider mode: `LLM_PROVIDER=fake` (no API key needed)
   - Real provider mode: `LLM_PROVIDER=openai` with API key
   - Step-by-step manual demo flow:
     1. Create task
     2. Create agent profiles
     3. Create agent team
     4. Add team members in order
     5. Create agent run from task
     6. Start run
     7. Monitor execution
     8. Cancel if needed
   - Full environment variables reference
   - Known limitations clearly documented
   - Troubleshooting section

**Tests Added**:
- Task context included in LLM messages
- Task loading failure marks run failed
- Empty TaskID skips task context gracefully

---

### Phase C.2.1: Read-Only Tool Execution MVP

**Status**: ✅ Complete

**Goal**: Give the orchestration engine "eyes" so agent runs can safely inspect the project workspace without modifying files.

**New Tool Package** (`services/api/internal/tool/`):

| File | Purpose |
|------|---------|
| `safety.go` | Path validation, blocked files/dirs, binary detection |
| `read_file.go` | Read text files with truncation + binary detection |
| `list_files.go` | List directory contents with depth control |
| `search_code.go` | Search code with plain text or regex |
| `registry.go` | Tool registration + definition metadata |
| `executor.go` | Tool call dispatch, JSON parsing, output formatting |

**Safety features**:
- Rejects absolute paths and `../` traversal
- Validates path stays within `WORKSPACE_ROOT`
- Blocks: `.env`, `.env.*`, `.git`, `node_modules`, `.next`, `dist`, `build`, `vendor`
- Binary detection by extension + null byte checking
- Symlink escape detection via `filepath.EvalSymlinks`
- Output size limits per tool

**Orchestrator Integration**:
- LLM can trigger tools via `{"tool_calls":[{"tool_name":"...","input":{...}}]}` format in response
- Tool calls recorded in `agent_tool_calls` with status `completed`/`failed`
- Tool results stored as `agent_messages` with `role: "tool"`
- One tool-execution pass per step (no feedback loop in MVP)
- Tool errors recorded gracefully (do not crash server)

**New Config** (`config.go`):
- `WORKSPACE_ROOT` — Project root (defaults to cwd)
- `TOOL_READ_MAX_BYTES` — Max bytes per file (default: 1MB)
- `TOOL_SEARCH_MAX_RESULTS` — Max search results (default: 50)

**Modified Files**:
- `config/config.go` — +3 tool env vars + `ToolConfig` struct
- `service/orchestrator.go` — +`CreateToolCall`/`UpdateToolCallStatus` in interface, tool execution in `executeRun()`
- `service/orchestrator_repo.go` — +`toolCallRepo`, +2 methods
- `server/router.go` — wire `toolCallRepo` to orchestrator, resolve workspace root
- `handler/agent_run_test.go` — updated mock for new interface methods
- `service/orchestrator_test.go` — updated mocks + 9 new tool integration tests

**Tests Added** (39 total):
- 12 safety/path validation tests
- 8 read_file tests
- 5 list_files tests
- 5 search_code tests
- 9 orchestrator tool integration tests
- All 390 backend tests passing

**Constraint**: Read-only only. No `write_file`, `edit_file`, `bash`, `git`, `apply_patch`.

---

### Phase C.2.2: Multi-Turn Tool Feedback Loop

**Status**: ✅ Complete

**Goal**: Enable the orchestration engine to iterate with the LLM — execute tool calls, feed results back, and let the LLM decide the next step.

**What Changed** (`services/api/internal/service/orchestrator.go`):

The `executeRun` function now loops per step:

```
for iter := 0; iter < MaxToolIterations; iter++ {
    1. Call LLM → response
    2. Store assistant message in agent_messages
    3. Append assistant message to LLM context
    4. Parse tool_calls from response
    5. If no tool_calls → final response, break
    6. Execute all tool calls
    7. Store tool calls in agent_tool_calls
    8. Store tool results as agent_messages (role: "tool")
    9. Append tool results to LLM context
    10. Continue loop
}
```

- Step output = final non-tool-call LLM response (the iteration where no `tool_calls` are found)
- If `MaxToolIterations` is reached without a non-tool-call response, falls back to `"Step completed."`
- Loop is cancel-safe: each iteration checks `ctx.Done()` before LLM call and after execution
- All messages stored in DB for observability

**New Config**:
- `TOOL_MAX_ITERATIONS` env var (default: 10) — added to `config.ToolConfig.MaxToolIterations`
- Wired through `tool.ToolOptions.MaxToolIterations` in `router.go`

**Test Updates** (`orchestrator_test.go`):
- Refactored `toolResponseLLM` → `newToolResponseLLM(responses ...string)` supporting variadic multi-turn responses
- Updated 3 integration tests to provide two-turn responses:
  1. `TestExecuteRun_ToolCallsCreateAgentToolCallRecords` — tool call → normal text
  2. `TestExecuteRun_ToolResultCreatesToolMessage` — tool call → normal text
  3. `TestExecuteRun_FailedToolCallDoesNotCrash` — unknown tool → normal text
- All 391 backend tests passing (+1 dedicated multi-turn feed verification test)

**Files Modified**:
- `services/api/internal/service/orchestrator.go` — multi-turn loop in `executeRun()` (core change)
- `services/api/internal/config/config.go` — `MaxToolIterations` in `ToolConfig`
- `services/api/internal/server/router.go` — wire `MaxToolIterations` into `tool.ToolOptions`
- `services/api/internal/tool/registry.go` — `MaxToolIterations` field in `ToolOptions`
- `services/api/internal/service/orchestrator_test.go` — refactored mocks + multi-turn responses + dedicated multi-turn feed verification test

**Key Design Decisions**:
- Full conversation history preserved: each LLM call gets `system + user + assistant + tool` messages
- Tool results appended to context immediately after tool execution
- Loop breaks on any non-tool-call response (LLM's natural final answer)
- Max iterations safety valve prevents runaway loops
- Single LLM response per iteration (no `n > 1`)

---

### Phase C.2.3: Tool Approval Gates

**Status**: ✅ Complete

**Goal**: Add pre-approval checks before tool execution — when an LLM requests a tool that requires human approval, the orchestrator creates a `human_approval` record and skips execution instead of running the tool immediately.

**What Changed**:

**Backend — Tool Options** (`tool/registry.go`):
- Added `RequireApproval []string` field to `ToolOptions`
- Added `RequiresApproval(toolName string) bool` method

**Backend — Config** (`config/config.go`):
- Added `TOOL_REQUIRE_APPROVAL` env var (comma-separated tool names, default empty)

**Backend — Router** (`server/router.go`):
- Parses comma-separated env var into `[]string`
- Wires into `tool.ToolOptions.RequireApproval`
- Passes `approvalRepo` to `NewOrchestratorRepoImpl`

**Backend — Orchestrator** (`service/orchestrator.go`):
- Adds `CreateApproval` to `OrchestratorRepository` interface
- In the multi-turn tool execution loop, before executing each tool:
  - If `RequiresApproval` → creates `human_approval` (status: `pending`) + `agent_tool_call` (status: `recorded`) + tool message ("requires human approval"); skips execution
  - Otherwise → executes tool normally

**Backend — Repo** (`service/orchestrator_repo.go`):
- Added `approvalRepo` field + `CreateApproval` implementation

**Frontend** (`HumanApprovalsPanel.tsx`, `AgentRunDetail.tsx`):
- Added `runStatus` prop to `HumanApprovalsPanel`
- Approvals panel polls every 3s while run is `running` to show auto-created approvals

**Tests Added** (3 new, 394 total):
1. `TestExecuteRun_ToolWithApprovalCreatesApprovalRecord` — verifies `recorded` status + approval message
2. `TestExecuteRun_ToolWithoutApprovalExecutesNormally` — verifies `completed` status + no approval message
3. `TestExecuteRun_ToolWithApprovalDoesNotFailRun` — verifies approval gate doesn't fail the run

**Files Modified**:
- `services/api/internal/tool/registry.go` — RequireApproval field + method
- `services/api/internal/config/config.go` — TOOL_REQUIRE_APPROVAL env var
- `services/api/internal/server/router.go` — parse + wire approval config
- `services/api/internal/service/orchestrator.go` — approval gate logic in loop
- `services/api/internal/service/orchestrator_repo.go` — approvalRepo integration
- `services/api/internal/service/orchestrator_test.go` — 3 new approval gate tests
- `services/api/internal/handler/agent_run_test.go` — mock CreateApproval method
- `apps/web/components/AgentRuns/HumanApprovalsPanel.tsx` — polling + runStatus prop
- `apps/web/components/AgentRuns/AgentRunDetail.tsx` — pass runStatus to panel

**Key Design Decisions**:
- Configurable per tool via `TOOL_REQUIRE_APPROVAL` env var
- Approval gate does NOT fail the step — LLM continues with feedback
- No automatic execution after approval (requires pause/resume, out of scope)
- Uses existing `human_approvals` table — no schema changes

---

### Phase C.2.4: Native LLM Function Calling

**Status**: ✅ Complete

**Goal**: Replace brittle JSON text parsing of tool calls with native OpenAI function calling API (`tools` parameter in chat completions), with backward-compatible text fallback for the FakeProvider.

**What Changed**:

**LLM Interface** (`llm/provider.go`):
- Added `ToolCall`, `ToolDefinition`, `ChatCompletionResponse` types
- Added `ChatCompletionWithTools(ctx, messages, tools)` to `LLMProvider` interface
- `ChatCompletion` remains for backward compatibility

**OpenAI Provider** (`llm/openai.go`):
- Added `openAITool`, `openAIFunc`, `openAIToolCall`, `openAIFuncCall` types
- Added `Tools` field to `openAIRequest` (uses `omitempty` so absent when no tools)
- Added `ToolCalls` field to `openAIMessage`
- `ChatCompletion` now delegates to `ChatCompletionWithTools` with nil tools (eliminates code duplication)
- `ChatCompletionWithTools`:
  - Converts `llm.ToolDefinition` → OpenAI `function` objects with `Type: "function"`
  - Sends `tools` parameter in request body
  - Parses `tool_calls` from response into `[]llm.ToolCall`
  - Returns `*ChatCompletionResponse` with both Content and ToolCalls

**Fake Provider** (`llm/fake.go`):
- `ChatCompletionWithTools` delegates to `ChatCompletion`, wraps result in `ChatCompletionResponse` (text-only, backward compat)

**Orchestrator** (`service/orchestrator.go`):
- Replaces `o.llm.ChatCompletion()` with `o.llm.ChatCompletionWithTools(ctx, messages, toolDefs)`
- Converts `tool.ToolDefinition` → `llm.ToolDefinition` before calling LLM
- Two tool call paths:
  1. **Native**: Uses `resp.ToolCalls` directly when LLM returns structured calls
  2. **Fallback**: Parses `resp.Content` with `tool.ParseToolCalls()` for backward compat
- Converts `llm.ToolCall` → `tool.ToolCallRequest` before executing
- Approval gate and normal execution logic unchanged

**Test Updates** (all mocks):
- Added `ChatCompletionWithTools` to `blockingFakeLLM`, `capturingLLM`, `toolResponseLLM`, `messagesRecorderLLM` (orchestrator tests)
- Added `ChatCompletionWithTools` to `mockLLM` (handler tests)
- All 394 tests pass with no behavioral changes

**Files Modified**:
- `services/api/internal/llm/provider.go` — new types + interface method
- `services/api/internal/llm/openai.go` — native function calling implementation
- `services/api/internal/llm/fake.go` — backward-compat wrapper
- `services/api/internal/service/orchestrator.go` — uses `ChatCompletionWithTools`
- `services/api/internal/service/orchestrator_test.go` — 4 mock LLM updates
- `services/api/internal/handler/agent_run_test.go` — mock LLM update

**Key Design Decisions**:
- No schema/migration changes
- Backward compatible: `ChatCompletion` still works (OpenAI delegates to `ChatCompletionWithTools` with nil tools)
- Fallback text parsing ensures FakeProvider continues working
- No `tool_choice` parameter exposed yet (LLM always receives all tools)

---

### Phase C.2.5: `tool_choice` Parameter

**Status**: ✅ Complete

**Goal**: Make the OpenAI `tool_choice` parameter configurable so users can control whether the LLM must call tools, may call tools, or must not call tools.

**What Changed**:

**Config** (`config/config.go`, `tool/registry.go`):
- Added `TOOL_CHOICE` env var to `ToolConfig` (default: `"auto"`)
- Added `ToolChoice string` to `ToolOptions` struct

**LLM Interface** (`llm/provider.go`):
- Added `toolChoice string` parameter to `ChatCompletionWithTools(ctx, messages, tools, toolChoice)`
- `ChatCompletion` passes `""` (no tool_choice) to maintain backward compat

**OpenAI Provider** (`llm/openai.go`):
- Added `ToolChoice any` field to `openAIRequest` with `omitempty`
- `ChatCompletionWithTools` validates toolChoice: only `"none"`, `"auto"`, `"required"` accepted
- Sets `reqBody.ToolChoice` when non-empty

**Fake Provider** (`llm/fake.go`):
- Updated signature; ignores `toolChoice` (text-based tool call parsing continues to work)

**Orchestrator** (`service/orchestrator.go`):
- Passes `o.toolOpts.ToolChoice` to `ChatCompletionWithTools`
- Router defaults ToolChoice to `"auto"` if empty

**Test Updates** — Updated 5 mock `ChatCompletionWithTools` signatures

**Files Modified**:
- `config/config.go` — `TOOL_CHOICE` env var
- `tool/registry.go` — `ToolChoice` in ToolOptions
- `llm/provider.go` — interface signature
- `llm/openai.go` — tool_choice in request
- `llm/fake.go` — signature update
- `service/orchestrator.go` — pass toolChoice
- `server/router.go` — default to "auto"
- `service/orchestrator_test.go` — 4 mocks
- `handler/agent_run_test.go` — 1 mock

**Verification**: `go build ./...` ✅, `go test ./...` ✅, 394 tests passing

---

### Phase C.3: Write Tool Execution

**Status**: ✅ Complete

**Goal**: Give the orchestration engine "hands" so agent runs can create/edit files, execute bash commands, and run git operations with comprehensive safety guards.

**New Files Created** (8):

| File | Purpose |
|------|---------|
| `services/api/internal/tool/write_file.go` | Create/overwrite files with safety: binary extension reject, null byte check, max size limit, parent dir auto-creation |
| `services/api/internal/tool/edit_file.go` | Find-and-replace in files: validates file exists, rejects binary, null byte check, no-match error, writes back modified content |
| `services/api/internal/tool/bash.go` | Shell command execution: timeout via context, 100KB output limit, workspace-root cwd, blocked command patterns |
| `services/api/internal/tool/git.go` | Git command execution: 60s timeout, destructive command blocking, output limits |
| `services/api/internal/tool/write_file_test.go` | 9 tests: create, overwrite, parent dirs, binary reject, dot-env reject, traversal reject, null bytes, size limit, empty path |
| `services/api/internal/tool/edit_file_test.go` | 9 tests: replace, multi-replace, no-match error, binary reject, traversal reject, non-existent file, empty path, empty old_string, dot-env reject |
| `services/api/internal/tool/bash_test.go` | 8 tests: execute echo, timeout, blocked command, workspace dir pwd, empty command, exit code, config-blocked, stderr capture |
| `services/api/internal/tool/git_test.go` | 10 tests: init, status, push-blocked, reset-blocked, clean-blocked, no-args, rebase-blocked, init+add+status, pull-blocked, merge-blocked |

**Files Modified** (6):

| File | Changes |
|------|---------|
| `services/api/internal/tool/safety.go` | Added `ValidateWritePath` (path exists check + anti-symlink-escape walk), `IsBashCommandAllowed` (40 blocked patterns), `IsGitCommandAllowed` (18 blocked commands + flag analysis) |
| `services/api/internal/tool/safety_test.go` | 22 new tests: 7 for ValidateWritePath, 5 for IsBashCommandAllowed, 14 for IsGitCommandAllowed |
| `services/api/internal/tool/registry.go` | Added `WriteMaxBytes`, `BashTimeout`, `BashBlocked` to ToolOptions; registered 4 new tool definitions and executor functions |
| `services/api/internal/config/config.go` | Added `WriteMaxBytes` (env: `TOOL_WRITE_MAX_BYTES`, default 1MB), `BashTimeout` (env: `TOOL_BASH_TIMEOUT`, default 30s), `BashBlocked` (env: `TOOL_BASH_BLOCKED_COMMANDS`) |
| `services/api/internal/server/router.go` | Passes new config to ToolOptions, added `parseBlockedCommands` helper |
| `services/api/internal/service/orchestrator_test.go` | Updated `TestToolRegistry_NoWriteToolsExist` → split into write tools exist + other tools still blocked |

**Safety Guards**:

| Domain | Protection |
|--------|------------|
| **Write path** | Absolute path reject, `../` traversal reject, `.env`/`.env.*` block, `.git`/`node_modules`/`.next`/`dist`/`build`/`vendor` block, symlink escape detection |
| **Binary content** | Extension-based reject (39 binary extensions) + null byte content check |
| **File size** | Configurable max write size via `TOOL_WRITE_MAX_BYTES` (default 1MB) |
| **Bash commands** | 40 blocked patterns: `sudo`, `su`, `shutdown`, `reboot`, `mkfs`, `fdisk`, `dd if=/of=`, `chown`, `apt`/`yum`/`dnf`/`brew`, redirect to `/dev/`/`/etc/`/`/proc/`/`/sys/`, pipe-to-shell (`| sh`, `| bash`), `curl ... | sh` |
| **Bash timeout** | Configurable via `TOOL_BASH_TIMEOUT` (default 30s), per-command override via input |
| **Bash output** | 100KB size limit for both stdout and stderr with truncation markers |
| **Git commands** | 18 blocked: `push`, `fetch`, `pull`, `rebase`, `clean`, `cherry-pick`, `merge`, `gc`, `prune`, `fsck`, `update-ref`, `reset` (all variants), `submodule update/deinit`, `tag --delete`, `config` writes to sensitive keys |
| **Git timeout** | Hardcoded 60s timeout |
| **Config extensibility** | `TOOL_BASH_BLOCKED_COMMANDS` env var for additional custom blocked bash patterns |

**Integration with Existing Orchestration**:
- Works transparently with the multi-turn tool feedback loop (C.2.2)
- Works transparently with approval gates (C.2.3) — set `TOOL_REQUIRE_APPROVAL=write_file,edit_file,bash,git` to require approval
- Works transparently with native LLM function calling (C.2.4)
- Works transparently with tool_choice parameter (C.2.5)
- No changes needed to orchestrator logic or data model
- No new migrations or schema changes

### Approval Auto-Resume (Phase C.3 Enhancement)

**Status**: ✅ Complete

**Goal**: When a human approves a tool call (via `PATCH /human-approvals/:id/status` → `"approved"`), the orchestrator should automatically execute the tool and feed the result into the LLM conversation without requiring manual re-run.

**Design**:
1. **Migration**: New column `tool_call_id` on `human_approvals` (FK to `agent_tool_calls`) — links each approval to its tool call
2. **Approval gate creation**: Reordered to create tool_call FIRST, then approval WITH the tool_call ID
3. **Auto-resume check**: Before each LLM call in the multi-turn loop, check for approved-but-not-executed tool calls
4. **Execution flow**:
   - Loop iteration N: LLM calls tool → approval gate → tool_call(recorded) + approval(pending, linked) → "requires approval" message → loop continues
   - Human approves externally → approval status = "approved"
   - Loop iteration N+1 (before LLM call): finds approved approval → executes tool → updates tool_call status to "completed" → creates tool message → appends to LLM context → calls LLM with result

**New files**:
| File | Purpose |
|------|---------|
| `db/migrations/000003_approval_tool_link.sql` | Add `tool_call_id` column + indexes |

**Modified files**:
| File | Changes |
|------|---------|
| `model/orchestration.go` | Added `ToolCallID *string` field to `HumanApproval` |
| `repository/human_approval.go` | Added `tool_call_id` to all SQL queries; new `GetApprovedToolApprovals()` method |
| `repository/agent_tool_call.go` | Added `UpdateOutput()` method (updates output + status) |
| `service/orchestrator.go` | Added `GetApprovedToolApprovals`/`UpdateToolCallOutput` to interface; reordered approval gate to link tool_call; added auto-resume check before each LLM call (lines 335-382) |
| `service/orchestrator_repo.go` | Delegation methods for new interface methods |
| `service/orchestrator_test.go` | New `autoResumeMockRepo` + `TestExecuteRun_AutoResumeExecutesApprovedToolCall` (verifies: run completes, tool executed, tool message created, LLM context contains result) |
| `handler/agent_run_test.go` | Mock methods for new interface methods |

**Test Results**:
- All backend: `go test ./...` → **All passing**
- Frontend: unaffected (54 tests passing, lint clean, build successful)

---

### C.3.2: Approvals Dashboard

**Status**: ✅ Complete

**Goal**: Add rich approval management UI: big Approve/Reject buttons for pending approvals, tool input/output display, auto-refresh after action.

**What Changed**:

| File | Changes |
|------|---------|
| `apps/web/lib/api.ts` | Added `tool_call_id?: string` to `HumanApproval` interface |
| `apps/web/lib/api.test.ts` | Added `tool_call_id: "tc1"` to mock approval |
| `apps/web/components/AgentRuns/HumanApprovalsPanel.tsx` | Major rewrite: approve/reject buttons with loading states, tool name extraction from `approval_type`, tool input as formatted JSON (expandable), tool call correlation via `tool_call_id`, tool result display (expandable), inline error handling, parallel fetch of approvals + tool calls on mount and polling |

**Key Features**:
- **✅ Approve button**: Green button, calls `PATCH /human-approvals/:id/status` with `"approved"`
- **❌ Reject button**: Red button, calls PATCH with `"rejected"`
- **Loading spinner**: Shows on clicked button, disables both buttons during API call
- **Dropdown (kept)**: For changing to "cancelled" status
- **Tool input**: Expandable `<details>` with formatted JSON from `request_notes`
- **Tool result**: For approved approvals with `tool_call_id`, shows tool call output in green `<pre>` block
- **Error handling**: Inline red banner on action failure
- **Polling**: Fetches both approvals AND tool calls every 3s while running

**Verification**:
- `npm run test:run` → 54 tests passing
- `npm run lint` → No warnings/errors
- `npm run build` → Compiles successfully (9 static pages)

### Phase C.4: Pause/Resume

**Status**: ✅ Complete

**Goal**: When the orchestrator encounters a tool that requires human approval, pause the run entirely instead of continuing. When the human approves, auto-resume and execute the approved tool.

**What Changed**:

**Backend — Orchestrator** (`service/orchestrator.go`):
- **Approval gate now PAUSES** the run instead of continuing with a "requires approval" message:
  1. Creates tool_call (status: "recorded")
  2. Creates approval (status: "pending", linked via tool_call_id)
  3. Marks step as `waiting_approval`
  4. Marks run as `paused`
  5. Exits the goroutine (defer cleans up the running map)
- **Added `ResumeRun` method**: Validates run is `paused`, atomically transitions to `running`, registers in running map, starts new `executeRun` goroutine
- **Step re-entry on resume**: When `executeRun` encounters a step with `waiting_approval`, it reloads all messages from DB to rebuild the LLM context and continues execution from where it left off
- **Auto-resume integration**: The existing auto-resume check (before each LLM call) finds the approved tool call and executes it, feeding the result into the LLM context
- Added `"waiting_approval"` to the pending step list so it's processed on resume

**Backend — Repository**:
- `agent_run_step.go`: Modified `MarkStarted` to use `COALESCE(started_at, NOW())` to preserve original timestamps; added `MarkWaitingApproval` method
- `agent_message.go`: Added `GetByStepID` method to reload messages on resume

**Backend — Handler**:
- `agent_run.go`: Added `ResumeAgentRun` handler → `POST /agent-runs/{id}/resume`
- `server/router.go`: Added resume route

**Frontend** (`apps/web/`):
- `lib/api.ts`: Added `resumeAgentRun()` API function
- `AgentRunDetail.tsx`: Added "▶ Resume Run" amber button for paused runs; extended polling to also run when paused; added `handleResumeRun` handler
- `HumanApprovalsPanel.tsx`: After successful approve/reject, auto-resumes the run if it's paused; extended polling to also run when paused

**Key Design Decisions**:
- Pause exits the goroutine entirely (no background polling for approval)
- Resume starts a fresh goroutine that re-enters `executeRun`
- Step messages are persisted in DB, so resume can rebuild LLM context from stored messages
- Auto-resume (from C.3.1) works transparently: resume → execute tool → feed result to LLM
- No new migrations needed (all statuses already existed in the enum)

**Tests Added/Updated**:
- `TestExecuteRun_ToolWithApprovalCreatesApprovalRecord` — now expects `paused` (not `completed`)
- `TestExecuteRun_ToolWithApprovalDoesNotFailRun` — now expects `paused` (not `completed`)
- `TestExecuteRun_AutoResumeExecutesApprovedToolCall` — restructured into two-phase test (pause → resume)
- `TestResumeRun_OnlyWorksForPaused` — validates resume succeeds for paused runs
- `TestResumeRun_FailsForNonPaused` — validates resume fails for non-paused statuses
- Handler test: `TestResumeAgentRun` — tests all resume handler scenarios

**Verification**:
- Backend: `go test ./...` → All passing (128+ test cases)
- Frontend: `npm run test:run` → 54 tests passing
- Frontend: `npm run lint` → No warnings/errors
- Frontend: `npm run build` → Success

---

## Latest Verification (After Phase C.4)

### Backend
- `go test ./...` → All passing (128+ test cases across all packages)
  - `internal/tool`: 83 tests
  - `internal/service`: All passing (15+ tests including pause/resume)
  - `internal/handler`: All passing (10+ tests including resume handler)
  - `internal/config`: All passing
  - `internal/server`: All passing
- LLM interface: `ChatCompletionWithTools` with native function calling + `tool_choice`
- OpenAI provider: sends `tools` API parameter, parses `tool_calls`, supports `tool_choice`
- FakeProvider: text parsing fallback (backward compatible)
- Tool package: `internal/tool/` with 7 registered tools, RequireApproval support
- Safety framework: path validation, binary detection, null byte checks, bash blocklist (40 patterns), git blocklist (18 commands), timeouts, output limits
- Multi-turn loop with approval gates + native tool calls in `internal/service/orchestrator.go`
- Write tool execution: write_file, edit_file, bash, git
- **Pause/Resume**: Approval gate pauses run; resume starts new goroutine; auto-resume executes approved tools
- No new migrations needed

### Frontend
- `npm run test:run` → 54 tests passing
- `npm run lint` → No warnings/errors
- `npm run build` → Compiles successfully (9 static pages)
- **Approvals Dashboard**: Approve/Reject buttons, tool input/output display, auto-refresh
- **Resume Button**: "▶ Resume Run" button when run is paused
- **Auto-resume**: Approving/rejecting a tool approval auto-resumes the paused run

### Database
- 2 migrations applied:
  - `000001_init.sql`
  - `000002_agent_orchestration.sql`
- 1 additive migration:
  - `000003_approval_tool_link.sql` (Phase C.3.1 — added `tool_call_id` to `human_approvals`)

### Key Endpoints Available
See `docs/CURRENT_STATUS.md` for full list.

### Key Phase C Endpoints
- `POST /agent-runs/:id/start` - Start run execution
- `POST /agent-runs/:id/cancel` - Cancel run execution
- `POST /agent-runs/:id/resume` - Resume paused run execution

### All Tools Available (7 total)
- **Read-only**: `list_files`, `read_file`, `search_code`
- **Write/Execute**: `write_file`, `edit_file`, `bash`, `git`

### Multi-Turn Tool Loop
- Iterative: LLM → tool calls → execute/approve → feed results back → re-query LLM → repeat
- Default max iterations: 10 (configurable via `TOOL_MAX_ITERATIONS`)
- Full conversation history preserved in each LLM call
- Approval gates: configurable per tool via `TOOL_REQUIRE_APPROVAL`
- **Pause/Resume**: When approval is required, run pauses; resume auto-executes approved tools
- Native function calling: OpenAI `tools` API with text fallback
- Write tools: binary rejects, size limits, blocked commands, timeouts
- **Approval auto-resume**: Approved tool calls are automatically executed and fed back to the LLM

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
