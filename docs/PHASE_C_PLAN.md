# Phase C Plan: Auto Orchestration MVP

## Phase C Title

**Auto Orchestration MVP** - Enable automatic execution of agent runs through an orchestration engine.

## Overview

Phase C builds on the existing database and API skeleton (Phase B.1) and the full UI (Phase B.2) to add actual execution capability. The goal is to have a working end-to-end system where:

1. User creates a task
2. User creates/selects an agent team
3. User starts an agent run
4. System auto-creates steps from team members
5. System executes steps in order
6. Dashboard shows real-time status via polling
7. All outputs (messages, tool calls, errors) are persisted

---

## C.1 Orchestrator Core

### Goal
Create the core orchestration logic that manages run and step state transitions.

### Key Components

1. **Orchestrator Service** (`internal/service/orchestrator.go`)
   - Main entry point for run execution
   - Manages step creation from team members
   - Handles step transitions based on dependencies
   - Tracks run status

2. **State Machine**
   - Run status transitions: `draft → pending → running → completed/failed/cancelled`
   - Step status transitions: `pending → running → completed/failed/cancelled/skipped`

3. **Step Factory**
   - Creates steps from team members
   - Maps `member_role` to step `step_type`
   - Uses `position` for ordering
   - Injects profile's `system_prompt` into step instructions

### Endpoints to Add
- `POST /agent-runs/:id/start` - Mark run as pending and trigger execution
- `POST /agent-runs/:id/cancel` - Set status to cancelled
- (Optional) `POST /agent-runs/:id/pause` / `POST /agent-runs/:id/resume`

### Files to Create/Modify
- `internal/service/orchestrator.go` (new)
- `internal/handler/agent_run.go` (add start/cancel handlers)
- `internal/repository/agent_run.go` (may need new queries)
- `internal/server/router.go` (add routes)

---

## C.2 LLM Provider Interface

### Goal
Create an abstraction layer for calling different LLM providers.

### Key Components

1. **LLM Provider Interface** (`internal/llm/provider.go`)
   ```go
   type Message struct {
       Role    string
       Content string
   }
   
   type LLMProvider interface {
       ChatCompletion(messages []Message) (string, error)
       ChatCompletionWithTools(messages []Message, tools []Tool) (string, []ToolCall, error)
   }
   ```

2. **Provider Implementations**
   - OpenAI-compatible API (for OpenAI, Anthropic via compatible endpoints, local models)
   - Configurable via environment variables

3. **Tool Definition Schema**
   - Standard format for describing tools to LLM
   - Maps to our internal tool call model

### Configuration
- Environment variables:
  - `LLM_PROVIDER` - e.g., "openai", "anthropic"
  - `LLM_API_KEY` - API key (from env, NOT hardcoded)
  - `LLM_MODEL` - Default model
  - `LLM_BASE_URL` - Optional for custom endpoints

### Files to Create/Modify
- `internal/llm/provider.go` (new)
- `internal/llm/openai.go` (new)
- `internal/config/config.go` (add LLM config)

---

## C.3 Execution Loop

### Goal
Create a simple execution loop that processes runs and steps.

### Approach
For MVP, use a simple in-memory loop or database polling. No external queue required.

Options (simplest first):
1. **Synchronous execution** - `POST /start` blocks until done (not for production, but simplest MVP)
2. **Background goroutine** - Start a goroutine per run
3. **Database poller** - Periodically check for pending runs

### Step Execution Flow

For each step:
1. Load step + associated profile
2. Build prompt context:
   - Profile's system prompt
   - Step instructions
   - Run goal
   - Previous step outputs (if any)
3. Call LLM with tools available
4. Handle tool calls:
   - Record tool call in database
   - Execute tool (read file, bash, etc. - for MVP, may be stubbed)
   - Record output in database
5. Record LLM response as message
6. Update step status + output
7. Move to next step

### Tool Execution (Stubbed for MVP)
For Phase C MVP, tools can be:
- Recorded in database (already exists)
- Actual execution can be added in later phase
- OR: Implement basic read-only tools for demo

### Files to Create/Modify
- `internal/service/executor.go` (new - step executor)
- `internal/tool/registry.go` (new - tool registry)
- `internal/tool/impl/*.go` (new - tool implementations)

---

## C.4 Dashboard Controls

### Goal
Add UI controls to start, pause, resume, cancel runs.

### Components to Add

1. **Run Controls** in `AgentRunDetail.tsx`
   - "Start Run" button (only if status is `draft` or `pending`)
   - "Pause" button (only if running)
   - "Resume" button (only if paused)
   - "Cancel" button (if not completed/failed/cancelled)

2. **Status Indicators**
   - Visual indicator when run is executing
   - Step-by-step progress visualization
   - Error display if run fails

3. **Polling (Optional)**
   - Simple `setInterval` to refresh run data
   - OR: Manual refresh button

### API Functions to Add
- `startAgentRun(id: string)`
- `cancelAgentRun(id: string)`
- (Optional) `pauseAgentRun`, `resumeAgentRun`

### Files to Modify
- `apps/web/lib/api.ts` (add start/cancel functions)
- `apps/web/components/AgentRuns/AgentRunDetail.tsx` (add controls)

---

## C.5 Guardrails and Tests

### Goal
Ensure the system is safe, tested, and follows security best practices.

### Guardrails

1. **Approval Gates**
   - Before executing certain tools (bash, file write), check if approval is required
   - For MVP: All potentially destructive operations require approval

2. **Timeout Protection**
   - Maximum execution time per step
   - Maximum total execution time per run

3. **Error Recovery**
   - Failed steps should mark run as failed, not crash
   - Persist all state before/after each operation

4. **Secret Protection**
   - No API keys in logs
   - No secrets in database (except encrypted config if needed)
   - Environment variables only

### Tests to Add

1. **Backend Tests**
   - Orchestrator: Run status transitions
   - Orchestrator: Step creation from team
   - Executor: Step execution happy path
   - Executor: Error handling
   - LLM Provider: Interface contract tests

2. **Frontend Tests**
   - Start/cancel API functions
   - Run controls rendering based on status

### Files to Create/Modify
- `internal/service/orchestrator_test.go` (new)
- `internal/service/executor_test.go` (new)
- `apps/web/lib/api.test.ts` (add start/cancel tests)

---

## Acceptance Criteria for Phase C

Phase C is complete when:

1. **POST /agent-runs/:id/start** starts execution
   - Run status transitions from `draft` → `pending` → `running`
   - Steps are auto-created from team members (if team is assigned)

2. **Step statuses transition correctly**
   - `pending` → `running` when execution starts
   - `running` → `completed` on success
   - `running` → `failed` on error
   - Position determines execution order

3. **Run status transitions correctly**
   - All steps completed → run `completed`
   - Any step failed → run `failed`
   - Cancel requested → run `cancelled`

4. **Messages/output/error are persisted**
   - LLM responses stored in `agent_messages`
   - Tool calls stored in `agent_tool_calls`
   - Step outputs stored in `agent_run_steps.output`
   - Errors visible in step/run status

5. **Dashboard updates**
   - User can see current status
   - User can start/cancel runs
   - (Optional) Polling auto-refreshes status

6. **No secrets hardcoded**
   - All API keys from environment variables
   - No keys in git-tracked files
   - .env files in .gitignore

7. **All tests/build pass**
   - Backend: `go test ./...`
   - Frontend: `npm run test:run`, `npm run lint`, `npm run build`

---

## Phase C Implementation Order

1. **C.1 Orchestrator Core** - State transitions + step creation
2. **C.2 LLM Provider Interface** - Abstraction + basic implementation
3. **C.3 Execution Loop** - Step execution + tool stubs
4. **C.4 Dashboard Controls** - Start/cancel UI
5. **C.5 Guardrails and Tests** - Tests + safety

---

## Out of Scope for Phase C

- WebSocket real-time updates (polling is fine)
- Distributed queue / Redis
- Authentication
- Actual tool execution (can be stubbed; read-only tools optional)
- Pause/resume (start/cancel sufficient for MVP)
- Retry logic for failed steps
