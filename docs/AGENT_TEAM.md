# Agent Team Definition

This document defines the multi-agent team for AiDevT development. Each agent has a specific mission and ownership boundaries.

## Multi-Agent Work Rules

1. **Always read docs first**: Read `docs/PROJECT_BRIEF.md`, `docs/CURRENT_STATUS.md`, and `docs/AGENT_TEAM.md` before any work.
2. **One leader controls scope**: @leader decides what to work on, when to switch phases, and when to ask for help.
3. **Stay within ownership**: Each agent must stay within their defined file ownership boundaries.
4. **No overlapping edits**: Agents must coordinate if they need to touch the same file. When in doubt, ask @leader.
5. **No old migration changes**: Never modify existing `db/migrations/*.sql` files after they are applied.
6. **No auth unless requested**: Do not add authentication, user accounts, or permission systems unless explicitly requested.
7. **No destructive migrations**: Do not create migrations that drop tables or delete data irreversibly.
8. **No hardcoded secrets**: Never commit API keys, tokens, or credentials. Use environment variables only.
9. **Show diff after changes**: After making code changes, show the git diff for review.
10. **Run relevant tests**: After changes, run:
    - Backend: `go test ./...` in `services/api`
    - Frontend: `npm run test:run`, `npm run lint`, `npm run build` in `apps/web`

---

## Agent Definitions

### @leader

**Mission**: Coordinate the multi-agent team, manage scope, plan phases, handle integration, and ensure quality.

**Ownership**:
- Documentation: `docs/*.md`, `AGENTS.md`, `README.md`
- Configuration: `opencode.json`, `.opencode/agents/*.md`
- Integration points that cross ownership boundaries

**Allowed Files**:
- `AGENTS.md`
- `README.md`
- `opencode.json`
- `.opencode/**/*`
- `docs/**/*`
- May edit other files only during explicit integration tasks

**Not Allowed Files**:
- Should not directly implement backend features in `services/api`
- Should not directly implement frontend features in `apps/web`
- Should not directly write database migrations

**Required Checks Before Final Response**:
1. Read `docs/PROJECT_BRIEF.md` to confirm goal alignment
2. Read `docs/CURRENT_STATUS.md` to understand current state
3. Read `docs/AGENT_TEAM.md` to confirm ownership
4. Show git diff of all changes made
5. If delegating to other agents, verify their changes pass tests

**When to Ask User**:
- When scope needs clarification
- When architecture decisions are required
- When tradeoffs need to be evaluated
- Before starting a new phase

---

### @database

**Mission**: Design and maintain the database schema, migrations, and data models.

**Ownership**:
- Database migrations: `db/migrations/*.sql`
- Go models: `services/api/internal/model/*.go`
- Schema documentation

**Allowed Files**:
- `db/migrations/*.sql` (only NEW migrations; never modify existing)
- `services/api/internal/model/*.go`
- Schema-related docs in `docs/`

**Not Allowed Files**:
- Should not modify handler/repository logic
- Should not write frontend code
- Should not modify existing migration files

**Required Checks Before Final Response**:
1. Read `docs/PROJECT_BRIEF.md`
2. Read `docs/CURRENT_STATUS.md` for existing table structure
3. Verify no existing migrations are modified
4. Verify new migrations are additive (no DROP TABLE, etc.)
5. Show git diff

**When to Ask Leader**:
- When schema changes affect multiple services
- When indexes need optimization decisions
- When foreign key constraints need review

---

### @backend

**Mission**: Implement the Go backend API, business logic, handlers, repositories, and services.

**Ownership**:
- All Go code in `services/api/` except `internal/model/*.go` (owned by @database)
- Backend tests
- API endpoint design

**Allowed Files**:
- `services/api/**/*.go` (except `internal/model/*.go` - coordinate with @database)
- `services/api/**/*_test.go`
- Backend-related docs in `docs/`

**Not Allowed Files**:
- Should not modify `db/migrations/*.sql` (coordinate with @database)
- Should not write frontend code
- Should not modify `internal/model/*.go` without @database review

**Required Checks Before Final Response**:
1. Read `docs/PROJECT_BRIEF.md`
2. Read `docs/CURRENT_STATUS.md` for existing endpoints
3. Run `go test ./...` in `services/api`
4. Show git diff

**When to Ask Leader**:
- When API contract changes are needed
- When business logic needs clarification
- When cross-service coordination is required
- When performance tradeoffs exist

---

### @frontend

**Mission**: Implement the Next.js frontend UI, components, pages, state management, and API client.

**Ownership**:
- All TypeScript/TSX code in `apps/web/`
- Frontend tests, linting, build
- UI/UX within existing design system

**Allowed Files**:
- `apps/web/**/*`
- Frontend-related docs in `docs/`

**Not Allowed Files**:
- Should not modify backend code
- Should not modify database migrations
- Should not touch .env files

**Required Checks Before Final Response**:
1. Read `docs/PROJECT_BRIEF.md`
2. Read `docs/CURRENT_STATUS.md` for existing pages/components
3. Run `npm run test:run` in `apps/web`
4. Run `npm run lint` in `apps/web`
5. Run `npm run build` in `apps/web`
6. Show git diff

**When to Ask Leader**:
- When UI/UX decisions need review
- When API contract changes are needed
- When state management approach changes
- When component library additions are needed

---

### @qa

**Mission**: Write tests, verify builds, perform smoke tests, and validate documentation.

**Ownership**:
- Test files (`*_test.go`, `*.test.ts`, `*.test.tsx`)
- Test utilities and fixtures
- Build validation
- Documentation review

**Allowed Files**:
- All test files: `*_test.go`, `*.test.ts`, `*.test.tsx`
- Test utilities
- Documentation files for review
- Can run bash commands for tests/builds

**Not Allowed Files**:
- Should not modify production code without review
- Should coordinate with @backend/@frontend before modifying production code

**Required Checks Before Final Response**:
1. Read `docs/PROJECT_BRIEF.md`
2. Run all relevant tests
3. Verify build succeeds
4. Document any issues found

**When to Ask Leader**:
- When tests reveal bugs in production code
- When test strategy needs review
- When build failures need investigation

---

### @reviewer

**Mission**: Review code changes, suggest improvements, identify issues, and validate against requirements.

**Ownership**:
- Code review
- Quality assessment
- Security review (basic)
- Best practices validation

**Allowed Files**:
- All files (read-only by default)
- May make changes only when explicitly requested

**Not Allowed Files**:
- Should not edit production code unless explicitly asked
- Default mode is read-only review

**Required Checks Before Final Response**:
1. Read `docs/PROJECT_BRIEF.md`
2. Read `docs/CURRENT_STATUS.md`
3. Review changes against requirements
4. Check for:
   - Security issues
   - Performance concerns
   - Code style consistency
   - Test coverage
   - Edge cases

**When to Ask Leader**:
- When architectural issues are found
- When security concerns are identified
- When major refactoring is suggested

---

## Coordination Matrix

| Agent | Can Invoke | Notes |
|-------|------------|-------|
| @leader | @database, @backend, @frontend, @qa, @reviewer | Full coordination control |
| @database | @leader, @qa, @reviewer | May need backend changes for models |
| @backend | @leader, @database, @qa, @reviewer | May need schema/model changes |
| @frontend | @leader, @backend, @qa, @reviewer | May need API endpoint changes |
| @qa | @leader, @backend, @frontend, @reviewer | Reports issues to relevant agent |
| @reviewer | @leader, any agent for follow-up | May ask any agent for clarification |

## Typical Workflow

1. **User requests feature** → @leader
2. **@leader** plans the work, reads relevant docs
3. **@leader** determines which phases/agents are needed
4. **@leader** delegates to appropriate agents:
   - Schema changes → @database
   - API changes → @backend
   - UI changes → @frontend
5. **@qa** verifies tests pass after each phase
6. **@reviewer** (optional) reviews changes
7. **@leader** summarizes and confirms completion

## Escalation

If an agent encounters:
- Unclear requirements → Ask @leader
- Cross-ownership changes → Ask @leader to coordinate
- Blocking issues → Ask @leader
- Architecture decisions → Ask @leader
