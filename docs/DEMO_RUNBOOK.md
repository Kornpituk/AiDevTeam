# AiDevT Demo Runbook

## Prerequisites

- Docker / Docker Compose
- Node.js 18+
- Go 1.22+

## Quick Start: Fake Provider Mode (No API Key Required)

Use `LLM_PROVIDER=fake` for local demo without an API key. The fake provider returns deterministic mock responses.

### 1. Start Postgres

```bash
docker-compose up -d postgres
```

### 2. Apply Migrations

```bash
cd services/api
go run cmd/api/main.go migrate
```

Or use any migration tool to run the SQL files in `db/migrations/`.

### 3. Configure Environment

Create a `.env` file in the project root:

```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/aidevt
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080

# Fake provider - no API key needed
LLM_PROVIDER=fake
```

### 4. Start Backend

```bash
cd services/api
go run cmd/api/main.go
```

Backend will be available at `http://localhost:8080`.

### 5. Start Frontend

```bash
cd apps/web
npm install
npm run dev
```

Frontend will be available at `http://localhost:3000`.

---

## Real Provider Mode (OpenAI API Key Required)

### Configure Environment

```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/aidevt
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080

# OpenAI provider
LLM_PROVIDER=openai
LLM_API_KEY=sk-your-api-key-here
LLM_MODEL=gpt-4-turbo-preview

# Optional: Custom timeout
# LLM_TIMEOUT=60s
```

---

## Manual Demo Flow

### 1. Create a Task

1. Go to `http://localhost:3000/tasks`
2. Click "+ New Task"
3. Fill in:
   - **Title**: "Implement User Authentication"
   - **Description**: "Add user login and registration functionality to the application"
   - **Status**: `pending`
4. Click "Create Task"
5. Note the Task ID from the URL or page

### 2. Create Agent Profiles

1. Go to `http://localhost:3000/agents`
2. Click "+ New Profile"
3. Create **Planner Agent**:
   - **Name**: "AI Planner"
   - **Role**: `planner`
   - **Description**: "Creates detailed implementation plans"
   - **System Prompt**: "You are an expert software planner. Analyze tasks and create detailed implementation plans. Break down work into clear, actionable steps."
   - **Default Model**: `gpt-4`
4. Click "Create Profile"
5. Create **Implementer Agent**:
   - **Name**: "AI Implementer"
   - **Role**: `implementer`
   - **Description**: "Writes code based on plans"
   - **System Prompt**: "You are an expert software developer. Implement features based on provided plans. Write clean, maintainable code."
   - **Default Model**: `gpt-4`
6. Click "Create Profile"
7. Create **Reviewer Agent**:
   - **Name**: "AI Reviewer"
   - **Role**: `reviewer`
   - **Description**: "Reviews code for quality"
   - **System Prompt**: "You are an expert code reviewer. Review implementations for quality, correctness, and best practices."
   - **Default Model**: `gpt-4`

### 3. Create an Agent Team

1. On the Agents page, click "+ New Team"
2. Fill in:
   - **Name**: "Dev Team Alpha"
   - **Description**: "Full-stack development team"
3. Click "Create Team"
4. Note the Team ID

### 4. Add Team Members in Order

1. On the Team Detail page, add members in execution order:
2. Add **AI Planner** first (Position 1, Role `planner`):
   - Click "+ Add Member"
   - Select "AI Planner" from dropdown
   - Member Role: `planner`
   - Position: `1`
   - Click "Add Member"
3. Add **AI Implementer** (Position 2, Role `implementer`):
   - Click "+ Add Member"
   - Select "AI Implementer" from dropdown
   - Member Role: `implementer`
   - Position: `2`
   - Click "Add Member"
4. Add **AI Reviewer** (Position 3, Role `reviewer`):
   - Click "+ Add Member"
   - Select "AI Reviewer" from dropdown
   - Member Role: `reviewer`
   - Position: `3`
   - Click "Add Member"

### 5. Create an Agent Run

1. Go to your Task Detail page (`/tasks/[task-id]`)
2. Or go to Tasks list and click on your task
3. Click "Create Agent Run" (if available) or create via API
4. Fill in:
   - **Team**: Select "Dev Team Alpha"
   - **Goal**: "Implement user authentication with login and registration"
   - **Status**: `draft`
5. Click "Create Run"
6. Note the Run ID

### 6. Start the Run

1. On the Run Detail page (`/agent-runs/[run-id]`):
2. Verify:
   - Status is `draft` or `planned`
   - Team is "Dev Team Alpha"
   - Goal is visible
3. Click **"Start Run"** button
4. Watch the execution:
   - Status changes to `running`
   - Polling refreshes every 3 seconds
   - Steps are created from team members (if not already created)
   - Each step executes in order

### 7. Monitor Execution

On the Run Detail page:

- **Run Status**: Shows current status (running, completed, failed, cancelled)
- **Summary**: Shows final summary when complete
- **Steps Panel**:
  - Shows each step's status (pending, running, completed, failed)
  - Shows step output when complete
- **Messages Panel**:
  - Shows LLM messages from each step
  - Click "Refresh" to see new messages
- **Approvals Panel**: Shows any approval gates
- **Tool Calls Panel**: Shows tool call records

### 8. Cancel a Running Run (Optional)

1. While run is `running`, click **"Cancel Run"** button
2. Run status changes to `cancelled`
3. Summary shows "Run cancelled."

---

## Environment Variables Reference

### Database

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | Postgres connection string | - |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `aidevt` |

### LLM

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | Provider type: `openai`, `fake` | `openai` |
| `LLM_API_KEY` | API key for OpenAI | - |
| `LLM_MODEL` | Model name | `gpt-4` |
| `LLM_BASE_URL` | Custom API base URL | OpenAI default |
| `LLM_TIMEOUT` | Request timeout (Go duration) | `60s` |

### Server

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Backend port | `8080` |
| `NEXT_PUBLIC_API_BASE_URL` | Frontend API URL | - |

---

## Known Limitations

### MVP Limitations (Intentional)

1. **In-Memory Execution Only**
   - Runs execute in goroutines
   - Runs do NOT survive server restart
   - No distributed queue (Redis/RabbitMQ)

2. **No WebSocket**
   - Dashboard uses polling (3-second interval)
   - Manual "Refresh" button available in Messages panel
   - No real-time push notifications

3. **No Real Dangerous Tool Execution**
   - Tool calls are database records only
   - No actual file write, bash, or git operations
   - Safe for demo without risk

4. **No Automatic Codebase Modification**
   - AI does NOT modify your files automatically
   - All execution must be explicitly planned, scoped, and approved
   - Phase C focuses on orchestration, not actual code changes

5. **No Authentication**
   - No user accounts
   - No login/OAuth/JWT
   - No permission systems

6. **Single Workspace Only**
   - No multi-tenant
   - Single project focus

### API Limitations

- `pause`/`resume` endpoints not implemented
- Only `start` and `cancel` available for run control

---

## Troubleshooting

### Backend Won't Start

1. Check Postgres is running:
   ```bash
   docker-compose ps
   ```

2. Check database connection:
   ```bash
   psql postgresql://postgres:postgres@localhost:5432/aidevt
   ```

3. Verify migrations applied:
   ```bash
   cd services/api
   go run cmd/api/main.go migrate
   ```

### LLM Issues

1. **Fake provider**: No network calls, should always work
2. **OpenAI provider**:
   - Verify `LLM_API_KEY` is set
   - Check network connectivity
   - Verify model availability

### Polling/Refresh Issues

1. Check run status is `running` (polling only runs while status is `running`)
2. Manual refresh:
   - Refresh browser page
   - Or click "Refresh" button in Messages panel

### Tests

Run backend tests:
```bash
cd services/api
go test ./...
```

Run frontend tests:
```bash
cd apps/web
npm run test:run
npm run lint
npm run build
```

---

## Docker Compose Reference

```yaml
# docker-compose.yml (simplified)
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: aidevt
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```
