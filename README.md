# git-bug-ax

An agent-centric CLI for [git-bug](https://github.com/MichaelMure/git-bug), providing the optimal "Agent Experience" for coordinating swarms of coding agents.

## Overview

`git-bug` is a decentralized, git-native issue tracker that stores issues as CRDT operations. `git-bug-ax` extends it with an alternate CLI optimized for autonomous coding agents rather than humans.

Key design goals:

- **Machine-first interface**: JSON responses with structured fields, prose only in descriptions
- **Coordination primitives**: Fine-grained status, dependency tracking, atomic claiming
- **Swarm-safe**: Works with git-bug's CRDT model and Lamport clock ordering
- **Human-compatible**: Issues remain viewable via `git-bug bug show`

## Architecture

### Storage Layers

| Layer | Contents | Purpose |
|-------|----------|---------|
| **Metadata** | status, type, priority, parent, blocks, claimed_by, required_capabilities | Coordination, querying, filtering |
| **Body (Markdown)** | scope, files-affected, implementation, acceptance criteria | Execution context |
| **Comments** | work log, status change events | Append-only audit trail |

### CRDT Considerations

git-bug stores issues as a series of operations in a CRDT, synchronized via git. The "view" of an issue at any point is a snapshot assembled from those operations.

- **Lamport clocks** order concurrent writes
- **Last-write-wins** semantics resolve conflicts
- **Single-direction relationships** (`parent`, `blocks`) prevent corruption; reverse indexes (`children`, `blocked_by`) are computed at snapshot assembly

Agent coordination pattern:

1. Agent claims task (writes status op)
2. Agent syncs/pulls
3. Agent verifies its claim won (checks current snapshot)
4. If claim lost → abort and pick different task
5. If claim held → proceed with work

### Metadata Namespacing

All ax-specific metadata fields are prefixed with `ax_` in the CRDT to avoid collision with git-bug's native fields. The prefix is stripped when returning snapshots to agents.

**Storage (CRDT):**
```
ax_status, ax_type, ax_priority, ax_parent, ax_blocks, ax_claimed_by
```

**Snapshot (API response):**
```
status, type, priority, parent, blocks, claimed_by
```

## Metadata Fields

### `ax_status`

Fine-grained status for agent coordination:

| Status | Meaning |
|--------|---------|
| `draft` | Created, not ready for work (planning incomplete) |
| `ready` | Unblocked, claimable |
| `claimed` | Agent has claimed, not yet started |
| `in-progress` | Active work underway |
| `blocked` | Waiting on dependency |
| `review` | Work complete, awaiting verification |
| `done` | Verified complete |
| `abandoned` | Agent gave up or crashed |
| `failed` | Attempted, couldn't satisfy acceptance criteria |
| `stale` | Claimed/in-progress too long without update |
| `needs-decomposition` | Too large, return to planning agent |
| `needs-replanning` | A failure suggests the implementation plan is flawed |
| `contested` | Detected concurrent claims |

The `abandoned` vs `failed` distinction matters—`abandoned` means "try again," `failed` means "needs human input or re-planning." A `needs-replanning` status triggers a feedback loop, automatically notifying a planning agent that the original approach was flawed and requires a new plan.

### `ax_type`

Issue classification for hierarchy and workflow:

- `epic` - Large body of work containing features/tasks
- `feature` - User-facing functionality
- `task` - Discrete unit of work
- `bug` - Defect fix
- `spike` - Research/investigation (no deliverable)
- `tech-debt` - Refactoring/cleanup

### `ax_priority`

Numeric priority for deterministic ordering. Lower numbers = higher priority.

Ordering for `next` and `claim-next`:

1. Sort by `priority` (ascending)
2. Tie-break by issue hash (lexicographic)

Since git-bug identifies issues by their initial hash, combining priority with hash guarantees deterministic ordering across all agents in a swarm—no two agents will disagree on which task is "next."

### `ax_required_capabilities`

An array of strings specifying agent capabilities required for the task (e.g., `["go", "react", "database"]`). This allows for intelligent task routing in a swarm with specialized agents.

### `ax_parent`

Reference to parent issue ID. Enables hierarchy:

```
epic-1
├── feature-1
│   ├── task-1
│   └── task-2
└── feature-2
```

### `ax_blocks`

Array of issue IDs that this issue blocks. Stored as forward reference only.

**Stored:** `task-1.blocks = ["task-2", "task-3"]`
**Computed:** `task-2.blocked_by = ["task-1"]`

### `ax_claimed_by`

Agent identifier of current claim holder. `null` when unclaimed.

## Placeholder References

When creating related issues, the issue hash isn't known until after creation. This creates a chicken-and-egg problem for `parent` and `blocks` relationships.

### Solutions

**1. Two-phase creation**

Create issues first, then update relationships:

```bash
ax create "feat(auth): add token validation" --type=task
# returns hash: d4e5f6

ax create "feat(auth): add refresh tokens" --type=task
# returns hash: g7h8i9

ax block g7h8i9 d4e5f6
ax reparent d4e5f6 a1b2c3
ax reparent g7h8i9 a1b2c3
```

**2. Symbolic names**

Assign a persistent symbolic name at creation that can be referenced before or after the hash is known:

```bash
ax create "feat(auth): JWT authentication" --type=epic --name=auth-epic
ax create "feat(auth): add token validation" --type=task --parent=@auth-epic
ax create "feat(auth): add refresh tokens" --type=task --parent=@auth-epic --blocks=@auth-task1
```

Symbolic names are stored in `ax_name` metadata and can be used interchangeably with hashes. The `@` prefix distinguishes names from hashes.

**3. Batch creation with placeholders**

Single operation creates multiple issues with temporary placeholders:

```bash
ax create-batch <<EOF
{
  "issues": [
    {"ref": "$epic", "title": "feat(auth): JWT authentication", "type": "epic"},
    {"ref": "$task1", "title": "feat(auth): add token validation", "type": "task", "parent": "$epic"},
    {"ref": "$task2", "title": "feat(auth): add refresh tokens", "type": "task", "parent": "$epic", "blocks": ["$task1"]}
  ]
}
EOF
```

Returns resolved IDs:

```json
{
  "$epic": "a1b2c3",
  "$task1": "d4e5f6",
  "$task2": "g7h8i9"
}
```

**4. Pending references**

Allow unresolved references that get resolved lazily:

```bash
ax create "feat(auth): add token validation" --type=task --parent=@auth-epic
# Warning: @auth-epic not found, relationship pending

ax create "feat(auth): JWT authentication" --type=epic --name=auth-epic
# Resolves pending parent reference for d4e5f6
```

Pending references are stored and resolved when the target issue is created with a matching symbolic name.

### Recommendation

Support all approaches—agents may work incrementally across sessions:

- **Symbolic names** for stable, human-readable references
- **Batch creation** for efficiency when creating related issues together
- **Pending references** for resilience when creation spans multiple sessions
- **Two-phase** always works as a fallback

### Labels

git-bug already provides `labels` for ad-hoc categorization. Agents should use labels for:

- Domain tagging (`backend`, `frontend`, `api`)
- Urgency signals (`urgent`, `low-priority`)
- Special handling (`needs-review`, `security`, `breaking-change`)
- Agent hints (`small`, `medium`, `large` for estimated size)

## Issue Titles

Issue titles should follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```
feat(auth): add JWT token validation
fix(api): handle null response in user endpoint
refactor(db): extract connection pooling logic
chore(deps): update dependencies
```

Benefits:

- Direct mapping from issue → commit message
- Type prefix aligns with `ax_type` field
- Scope signals affected area
- Enables automated changelog generation
- Machine-parseable

| Issue type | Conventional prefix |
|------------|---------------------|
| feature | `feat:` |
| bug | `fix:` |
| tech-debt | `refactor:` or `chore:` |
| spike | `chore:` |
| epic | Aggregate description (children carry specific prefixes) |

## Body Format

The issue body uses Markdown with structured sections. Sections are both human-readable and machine-parseable.

### Canonical Sections

```markdown
## Scope

- Brief description of what this task accomplishes
- Boundaries of the work

## Files Affected

- pkg/api/handler.go
- pkg/validate/rules.go

## Environment

- Details for reproducing the development environment
- Required dependencies, environment variables, or secrets
- Link to a devcontainer definition or Dockerfile

## Implementation

- Use existing validator package
- Add new validation rules for email format
- Update handler to call validator before processing

## Acceptance Criteria

- All user inputs validated before processing
- Invalid inputs return 400 with descriptive error
- Unit tests cover new validation rules
- Existing tests pass

## Verification

- `go test ./pkg/api -run TestUserValidation`
```

### Parsing Rules

Parsing is lenient to accommodate human edits:

1. Find `## Section Name` headers (case-insensitive, fuzzy match aliases)
2. Extract content until next `##` or EOF
3. If content is a list → array of items
4. If content is prose → single string or array of paragraphs
5. Unrecognized sections → preserved in `sections.other`

### Section Aliases

| Canonical | Aliases |
|-----------|---------|
| `Scope` | `scope`, `Summary`, `summary` |
| `Files Affected` | `files-affected`, `Files`, `files` |
| `Implementation` | `implementation`, `Implementation Details` |
| `Acceptance Criteria` | `acceptance criteria`, `AC`, `Done When` |
| `Environment` | `environment`, `env`, `Setup` |
| `Verification` | `verification`, `Test Plan`, `Test Command` |

### Validation

```bash
ax validate <id>
```

Returns warnings (not errors) for malformed bodies:

```json
{
  "valid": true,
  "warnings": [
    "Section 'Acceptance Criteria' not found",
    "Section 'Scope' contains prose, expected list"
  ]
}
```

### Normalization

```bash
ax normalize <id>
```

Rewrites body with canonical headers and formatting while preserving content.

## Snapshot Format

All operations return JSON snapshots:

```json
{
  "id": "abc123",
  "title": "feat(auth): add JWT token validation",
  "type": "task",
  "status": "ready",
  "priority": 10,
  "parent": "epic-456",
  "blocks": [],
  "blocked_by": ["task-789"],
  "children": [],
  "claimed_by": null,
  "labels": ["backend", "security"],
  "created": "2026-01-20T10:00:00Z",
  "updated": "2026-01-24T15:30:00Z",
  "body": {
    "raw": "## Scope\n- Add validation...",
    "sections": {
      "scope": ["Add JWT validation to auth middleware"],
      "files_affected": ["pkg/api/handler.go", "pkg/validate/rules.go"],
      "implementation": ["Use existing validator package..."],
      "acceptance_criteria": ["All inputs validated", "Tests pass"],
      "verification": ["go test ./pkg/api -run TestUserValidation"],
      "environment": ["Requires GO_API_KEY to be set"]
    }
  },
  "recent_log": [
    {"agent": "planner-01", "timestamp": "2026-01-20T10:00:00Z", "message": "Created from epic-456"}
  ]
}
```

## Operations

### Query Operations

Read-only operations that don't create CRDT ops.

| Operation | Description | Returns |
|-----------|-------------|---------|
| `ax ready` | Unblocked, claimable tasks | Array of snapshots |
| `ax ready --type=task` | Filter by type | Array of snapshots |
| `ax ready --label=backend` | Filter by label | Array of snapshots |
| `ax ready --has-capability=go` | Filter by agent capability | Array of snapshots |
| `ax mine` | Tasks claimed by this agent | Array of snapshots |
| `ax blocked` | Tasks waiting on dependencies | Array with blocker info |
| `ax children <id>` | Subtasks of an issue | Array of snapshots |
| `ax blockers <id>` | What blocks this task | Array of snapshots |
| `ax show <id>` | Full snapshot | Single snapshot |
| `ax files <path>` | Tasks affecting a file/directory | Array of snapshots |
| `ax next` | Highest priority ready task | Single snapshot |

### Mutation Operations

Operations that create CRDT ops. All mutations return the resulting snapshot.

| Operation | Description | CRDT Effect |
|-----------|-------------|-------------|
| `ax claim <id>` | Claim a task | Set status=claimed, claimed_by=agent |
| `ax start <id>` | Begin work | Set status=in-progress |
| `ax block <id> <blocker-id>` | Add dependency | Append to blocks list |
| `ax unblock <id> <blocker-id>` | Remove dependency | Remove from blocks list |
| `ax complete <id>` | Finish work | Set status=review or done |
| `ax abandon <id> [reason]` | Give up task | Set status=abandoned |
| `ax fail <id> <reason>` | Mark as failed | Set status=failed |
| `ax reparent <id> <parent-id>` | Change parent | Set parent field |
| `ax update-body <id> <section> <content>` | Edit body section | Replace Markdown section |
| `ax label <id> <label>` | Add label | Append to labels |
| `ax unlabel <id> <label>` | Remove label | Remove from labels |
| `ax priority <id> <value>` | Set priority | Set priority field |

### Compound Operations

Convenience operations combining multiple steps.

| Operation | Description | Effect |
|-----------|-------------|--------|
| `ax claim-next` | Claim highest priority ready task | Find + claim atomically |
| `ax claim-next --type=task` | Filter by type | Find + claim with filter |
| `ax claim-next --has-capability=go` | Filter by agent capability | Find + claim with filter |
| `ax log <id> <message>` | Add work log entry | Append comment |
| `ax verify-claim <id>` | Check claim status | Returns boolean + current snapshot |
| `ax decompose <id> <child-ids...>` | Create subtask relationship | Set children's parent, update original status |

## MCP Tools

All operations are exposed as MCP tools with identical semantics:

```json
{
  "name": "ax_ready",
  "parameters": {
    "type": "task",
    "label": "backend"
  }
}
```

```json
{
  "name": "ax_claim",
  "parameters": {
    "id": "abc123"
  }
}
```

## Future Considerations

### Agent Identity & Lifecycle

- [ ] Agent registration / naming convention (e.g., `planner-01`, `coder-swarm-03`, `qa-01`)
- [ ] Agent capability registration (e.g., `go`, `react`, `database`)
- [ ] Heartbeat or TTL on claims → auto-transition to `stale` after timeout
- [ ] Claim history (who attempted, not just current holder)

### Execution & Workflow

- [ ] `branch` field in metadata → naming convention or assigned branch for work
- [ ] Formalize handoffs, e.g., a `qa-agent` that automatically picks up `review` tasks, runs the `Verification` command, and moves to `done` or `needs-replanning`.

### Observability

- [ ] Metrics endpoint: claims/completions/failures per agent
- [ ] Contention tracking: which tasks get claimed multiple times

### Safety

- [ ] Max concurrent claims per agent
- [ ] Circuit breaker: pause agent if failure rate spikes

### Templates

- [ ] Per-type body templates so planning agents generate consistent structure

### Advanced Concepts: Trust, Knowledge, and Economics

For a more mature swarm, the following concepts provide a path to greater autonomy and intelligence:

- **Cryptographic Agent Identity & Trust**:
  - [ ] Agents generate public/private key pairs for identity.
  - [ ] All CRDT operations (claims, status changes) are cryptographically signed.
  - This provides non-repudiation, enables access control (e.g., only planners can create epics), and establishes a zero-trust model between agents.

- **Shared Knowledge Base**:
  - [ ] Implement a parallel, git-based knowledge base (e.g., Markdown files in a repo) for durable, swarm-wide learning.
  - [ ] Agents can write Architectural Decision Records (ADRs), post-mortems for failed tasks, or best-practice guides.
  - [ ] Issues can link to knowledge base articles, allowing agents to learn from past successes and failures, reducing redundant work.

- **Economic Primitives for Prioritization**:
  - [ ] Introduce `budget` and `bounty` metadata fields.
  - [ ] Epics/features are allocated a `budget`. Planners assign a `bounty` to each sub-task based on value/difficulty.
  - [ ] Agents "earn" bounties, creating a market-based incentive system that dynamically prioritizes the most valuable work and provides clear metrics for agent performance and cost control.

## Installation

TODO

## Usage

```bash
# Find available work
ax ready

# Claim highest priority task
ax claim-next --type=task

# Start work
ax start abc123

# Log progress
ax log abc123 "Implemented validation logic"

# Complete task
ax complete abc123
```

## License

TODO
